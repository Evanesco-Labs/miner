package miner

import (
	"errors"
	"fmt"
	"github.com/Evanesco-Labs/Miner/keypair"
	"github.com/Evanesco-Labs/Miner/problem"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"io"
	"sync"
)

var (
	ErrorMinerWorkerOutOfRange = errors.New("miner's workers reach MaxWorkerCnt, can not add more workers")
)

type TaskStep int

const (
	TASKSTART TaskStep = iota
	TASKWAITCHALLENGEBLOCK
	TASKGETCHALLENGEBLOCK
	TASKPROBLEMSOLVED
	TASKSUBMITTED
)

type Task struct {
	coinbase         common.Address
	Step             TaskStep
	lastCoinBaseHash [32]byte
	challengeHeader  *types.Header
	challengeIndex   Height
	lottery          *problem.Lottery
}

func (t *Task) SetHeader(h *types.Header) {
	t.challengeHeader = h
	t.lottery.ChallengHeaderHash = h.Hash()
	t.Step = TASKGETCHALLENGEBLOCK
}

//SetTaskCoinbase only use in TASKSTART step
func SetTaskCoinbase(template *Task, coinbase common.Address) Task {
	if template.Step != TASKSTART {
		panic("only use it to update task in step TASKSTART")
	}
	return Task{
		coinbase:         coinbase,
		Step:             TASKSTART,
		lastCoinBaseHash: template.lastCoinBaseHash,
		challengeIndex:   Height(uint64(0)),
		lottery: &problem.Lottery{
			Address: coinbase,
		},
	}
}

type Config struct {
	CoinbaseList     []keypair.Key
	MaxWorkerCnt     int32
	MaxTaskCnt       int32
	CoinbaseInterval uint64
	SubmitAdvance    uint64
	//Test
	HeaderCh chan *types.Header
	//Test
	getHeader func(height Height) (*types.Header, error)
}

type Miner struct {
	mu               sync.RWMutex
	config           Config
	zkpProver        *problem.ProblemProver
	MaxWorkerCnt     int32
	MaxTaskCnt       int32
	workers          map[common.Address]*Worker
	scanner          *Scanner
	coinbaseInterval Height
	submitAdvance    Height
	exitCh           chan struct{}

	//todo: add a block header source
}

func NewMiner(config Config, r1cs io.Reader, pk io.Reader) (*Miner, error) {
	zkpProver, err := problem.NewProblemProver(r1cs, pk)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	miner := Miner{
		mu:               sync.RWMutex{},
		config:           config,
		zkpProver:        zkpProver,
		MaxWorkerCnt:     config.MaxWorkerCnt,
		MaxTaskCnt:       config.MaxTaskCnt,
		workers:          make(map[common.Address]*Worker),
		coinbaseInterval: Height(config.CoinbaseInterval),
		submitAdvance:    Height(config.SubmitAdvance),
		exitCh:           make(chan struct{}),
	}
	miner.StartScanner()
	go miner.Loop()
	//add new workers
	for _, key := range config.CoinbaseList {
		miner.NewWorker(key)
	}
	fmt.Println("start miner")
	return &miner, nil
}

func (m *Miner) Close() {
	//close workers
	for _, worker := range m.workers {
		worker.close()
	}
	//close scanner
	m.scanner.close()

	close(m.exitCh)
}

func (m *Miner) NewWorker(coinbase keypair.Key) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.workers) == int(m.MaxWorkerCnt) {
		fmt.Println(ErrorMinerWorkerOutOfRange.Error())
		log.Error(ErrorMinerWorkerOutOfRange.Error())
		return
	}
	worker := Worker{
		mu:               sync.RWMutex{},
		running:          0,
		MaxTaskCnt:       m.MaxTaskCnt,
		coinbase:         coinbase.Address,
		pk:               coinbase.PrivateKey.Public(),
		sk:               &coinbase.PrivateKey,
		workingTaskCnt:   0,
		coinbaseInterval: m.coinbaseInterval,
		inboundTaskCh:    make(chan *Task),
		submitAdvance:    m.submitAdvance,
		scanner:          m.scanner,
		zkpProver:        m.zkpProver,
		exitCh:           make(chan struct{}),
	}

	m.workers[coinbase.Address] = &worker
	go worker.Loop()
	worker.start()
	fmt.Println("worker start")
}

func (m *Miner) CloseWorker(coinbase common.Address) {
	if worker, ok := m.workers[coinbase]; ok {
		worker.close()
		delete(m.workers, coinbase)
	}
}

func (m *Miner) StopWorker(coinbase common.Address) {
	if worker, ok := m.workers[coinbase]; ok {
		worker.stop()
	}
}

func (m *Miner) StartWorker(coinbase common.Address) {
	if worker, ok := m.workers[coinbase]; ok {
		worker.start()
	}
}

func (m *Miner) Loop() {
	for {
		select {
		case <-m.exitCh:
			return
		case taskTem := <-m.scanner.outboundTaskCh:
			if taskTem.Step == TASKSTART {
				for _, worker := range m.workers {
					task := SetTaskCoinbase(taskTem, worker.coinbase)
					worker.inboundTaskCh <- &task
				}
				continue
			}
			if taskTem.Step == TASKGETCHALLENGEBLOCK {
				if worker, ok := m.workers[taskTem.coinbase]; ok {
					worker.inboundTaskCh <- taskTem
					continue
				}
				log.Warn("worker for this task not exist")
			}
			if taskTem.Step == TASKSUBMITTED {
				//todo store submitted lotteries
				log.Info("new task submitted")
			}
		}
	}
}

func (m *Miner) StartScanner() {
	m.scanner = &Scanner{
		mu:                 sync.RWMutex{},
		BestScore:          zero,
		LastBlockHeight:    0,
		CoinbaseInterval:   m.coinbaseInterval,
		LastCoinbaseHeight: 0,
		taskWait:           make(map[Height][]*Task),
		//todo:change this to real HeaderCh source
		//Test
		headerCh:       m.config.HeaderCh,
		inboundTaskCh:  make(chan *Task),
		outboundTaskCh: make(chan *Task),
		exitCh:         make(chan struct{}),
		getHeader:      m.config.getHeader,
	}

	go m.scanner.Loop()

	//todo: send header to scanner header channel
}
