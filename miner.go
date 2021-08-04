package miner

import (
	"github.com/Evanesco-Labs/Miner/problem"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"sync"
)

type Config struct {
	Coinbase common.Address
	Workers  int
}

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

type Miner struct {
	mu sync.RWMutex

	//block state source to new a scanner
	MaxWorkerCnt int
	workers      map[common.Address]*Worker
	scanner      *Scanner
	taskToSubmit chan *Task
	startCh      chan struct{}
	stopCh       chan struct{}
	exitCh       chan struct{}
}

func (m *Miner) Loop() {
	for {
		select {
		case <-m.startCh:
		case <-m.stopCh:
		case <-m.exitCh:
		case taskTem := <-m.scanner.minerTaskCh:
			if taskTem.Step == TASKSTART {
				for _, worker := range m.workers {
					task := SetTaskCoinbase(taskTem, worker.coinbase)
					worker.minerTaskCh <- &task
				}
				continue
			}
			if taskTem.Step == TASKGETCHALLENGEBLOCK {
				if worker, ok := m.workers[taskTem.coinbase]; ok {
					worker.minerTaskCh <- taskTem
					continue
				}
				log.Warn("worker for this task not exist")
			}
		case task := <-m.taskToSubmit:
			err := m.scanner.Submit(task)
			if err != nil {
				log.Error(err.Error())
			}
		}
	}
}

type ZKPProver struct {
	r1cs frontend.CompiledConstraintSystem
	pk   groth16.ProvingKey
}

func (p *ZKPProver) Prove(preimage []byte) ([]byte, []byte) {
	return problem.ZKPProve(p.r1cs, p.pk, preimage)
}

func InitZKPProver() ZKPProver {
	r1cs := problem.CompileCircuit()
	pk, _ := problem.SetupZKP(r1cs)
	return ZKPProver{
		r1cs: r1cs,
		pk:   pk,
	}
}

type ZKPVerifier struct {
	r1cs frontend.CompiledConstraintSystem
	vk   groth16.VerifyingKey
}

func InitZKPVerifier() ZKPVerifier {
	r1cs := problem.CompileCircuit()
	_, vk := problem.SetupZKP(r1cs)
	return ZKPVerifier{
		r1cs: r1cs,
		vk:   vk,
	}
}
