package miner

import (
	"context"
	"errors"
	"github.com/Evanesco-Labs/Miner/keypair"
	"github.com/Evanesco-Labs/Miner/log"
	"github.com/Evanesco-Labs/Miner/problem"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"io"
	"sync"
	"time"
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
	CoinbaseAddr     common.Address
	minerAddr        common.Address
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

func (t *Task) SetCoinbaseAddr(coinbaseAddr common.Address) {
	t.CoinbaseAddr = coinbaseAddr
}

//SetTaskMinerAddr only use in TASKSTART step
func SetTaskMinerAddr(template *Task, minerAddr common.Address) Task {
	if template.Step != TASKSTART {
		panic("only use it to update task in step TASKSTART")
	}
	//Deep Copy task
	return Task{
		minerAddr:        minerAddr,
		CoinbaseAddr:     template.CoinbaseAddr,
		Step:             TASKSTART,
		lastCoinBaseHash: template.lastCoinBaseHash,
		challengeIndex:   Height(uint64(0)),
		lottery: &problem.Lottery{
			MinerAddr:    minerAddr,
			CoinbaseAddr: template.CoinbaseAddr,
		},
	}
}

type Config struct {
	MinerList        []keypair.Key
	MaxWorkerCnt     int32
	MaxTaskCnt       int32
	CoinbaseInterval uint64
	SubmitAdvance    uint64
	CoinbaseAddr     common.Address
	WsUrl            string
	RpcTimeout       time.Duration
}

type Miner struct {
	mu               sync.RWMutex
	config           Config
	zkpProver        *problem.ProblemProver
	MaxWorkerCnt     int32
	MaxTaskCnt       int32
	CoinbaseAddr     common.Address
	workers          map[common.Address]*Worker
	scanner          *Scanner
	coinbaseInterval Height
	submitAdvance    Height
	wsUrl            string
	exitCh           chan struct{}
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
		CoinbaseAddr:     config.CoinbaseAddr,
		workers:          make(map[common.Address]*Worker),
		coinbaseInterval: Height(config.CoinbaseInterval),
		submitAdvance:    Height(config.SubmitAdvance),
		exitCh:           make(chan struct{}),
		wsUrl:            config.WsUrl,
	}

	err = miner.NewScanner()
	if err != nil {
		return nil, err
	}

	go miner.Loop()
	//add new workers
	for _, key := range config.MinerList {
		miner.NewWorker(key)
	}
	log.Info("miner start")
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

func (m *Miner) NewWorker(minerKey keypair.Key) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.workers) == int(m.MaxWorkerCnt) {
		log.Error(ErrorMinerWorkerOutOfRange.Error())
		return
	}
	worker := Worker{
		mu:               sync.RWMutex{},
		running:          0,
		MaxTaskCnt:       m.MaxTaskCnt,
		CoinbaseAddr:     m.CoinbaseAddr,
		minerAddr:        minerKey.Address,
		pk:               minerKey.PrivateKey.Public(),
		sk:               &minerKey.PrivateKey,
		workingTaskCnt:   0,
		coinbaseInterval: m.coinbaseInterval,
		inboundTaskCh:    make(chan *Task),
		submitAdvance:    m.submitAdvance,
		scanner:          m.scanner,
		zkpProver:        m.zkpProver,
		exitCh:           make(chan struct{}),
	}

	m.workers[minerKey.Address] = &worker
	go worker.Loop()
	worker.start()
	log.Debug("worker start")
}

func (m *Miner) CloseWorker(addr common.Address) {
	if worker, ok := m.workers[addr]; ok {
		worker.close()
		delete(m.workers, addr)
	}
}

func (m *Miner) StopWorker(addr common.Address) {
	if worker, ok := m.workers[addr]; ok {
		worker.stop()
	}
}

func (m *Miner) StartWorker(addr common.Address) {
	if worker, ok := m.workers[addr]; ok {
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
					task := SetTaskMinerAddr(taskTem, worker.minerAddr)
					worker.inboundTaskCh <- &task
				}
				continue
			}
			if taskTem.Step == TASKGETCHALLENGEBLOCK {
				if worker, ok := m.workers[taskTem.minerAddr]; ok {
					worker.inboundTaskCh <- taskTem
					continue
				}
				log.Warn("worker for this task not exist")
			}
			if taskTem.Step == TASKSUBMITTED {
				//todo: store submitted lotteries for later queries
			}
		}
	}
}

func (m *Miner) NewScanner() error {
	client, err := rpc.Dial(m.wsUrl)

	headerCh := make(chan *types.Header)
	sub, err := client.EthSubscribe(context.Background(), headerCh)

	if err != nil {
		return err
	}

	m.scanner = &Scanner{
		mu:                 sync.RWMutex{},
		CoinbaseAddr:       m.CoinbaseAddr,
		BestScore:          zero,
		LastBlockHeight:    0,
		CoinbaseInterval:   m.coinbaseInterval,
		LastCoinbaseHeight: 0,
		taskWait:           make(map[Height][]*Task),
		headerCh:           headerCh,
		inboundTaskCh:      make(chan *Task),
		outboundTaskCh:     make(chan *Task),
		exitCh:             make(chan struct{}),
		wsUrl:              m.wsUrl,
		evaClient:          client,
		sub:                sub,
		rpcTimeout:         m.config.RpcTimeout,
	}

	go m.scanner.Loop()
	return nil
}
