package miner

import (
	"errors"
	"github.com/Evanesco-Labs/Miner/keypair"
	"github.com/Evanesco-Labs/Miner/vrf"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"go.uber.org/atomic"
	"math/big"
	"sync"
)

var (
	InvalidStepError = errors.New("invalid task step")
	ZKPProofError    = errors.New("zkp prove error")
)

type Worker struct {
	mu               sync.RWMutex
	MaxTaskCnt       int32
	coinbase         common.Address
	pk               *keypair.PublicKey
	sk               *keypair.PrivateKey
	workingTaskCnt   atomic.Int32
	coinbaseInterval uint64
	minerTaskCh      chan *Task //channel to get task from miner
	solvedTaskCh     chan *Task //send finished task to miner
	//scannerTaskCh    chan *Task //channel to send waiting task to scanner
	scanner   *Scanner
	zkpProver *ZKPProver
	startCh   chan struct{}
	stopCh    chan struct{}
	exitCh    chan struct{}
}

func (w *Worker) Loop() {
	for {
		select {
		case <-w.startCh:
		case <-w.stopCh:
		case <-w.exitCh:
		case task := <-w.minerTaskCh:
			if task.Step == TASKSTART {
				err := w.HandleStartTask(task)
				if err != nil {
					log.Error(err.Error())
				}
				continue
			}
			if task.Step == TASKGETCHALLENGEBLOCK {
				err := w.HandleChallengedTask(task)
				if err != nil {
					log.Error(err.Error())
				}
				continue
			}
		}
	}
}

func (w *Worker) HandleStartTask(task *Task) error {
	task.coinbase = w.coinbase
	task.lottery.SetCoinbase(w.coinbase)
	index, proof := vrf.Evaluate(w.sk, task.lastCoinBaseHash[:])
	task.challengeIndex = Height(GetChallengeIndex(index, w.coinbaseInterval))
	task.lottery.VrfProof = proof
	task.lottery.Index = index
	task.Step = TASKWAITCHALLENGEBLOCK

	// request if this block already exit
	if w.scanner.LastBlockHeight+task.challengeIndex <= w.scanner.LastBlockHeight {
		return w.HandleTaskAfterChallenge(task)
	}

	// waiting for challenge block exist
	return w.HandlerTaskBeforeChallenge(task)
}

func (w *Worker) HandleChallengedTask(task *Task) error {
	// start zkp proof
	err := w.SolveProblem(task)
	if err != nil {
		return err
	}

	//submit lottery
	//give it to miner to submit
	err = w.SubmitTask(task)
	if err != nil {
		return err
	}
	w.solvedTaskCh <- task
	return nil
}

func (w *Worker) HandleTaskAfterChallenge(task *Task) error {
	header, err := w.scanner.GetHeader(w.scanner.LastBlockHeight + task.challengeIndex)
	//Drop this task if cannot get header
	//todo: maybe check if connection still remains
	if err != nil {
		return err
	}
	task.SetHeader(header)
	return w.HandleChallengedTask(task)
}

func (w *Worker) HandlerTaskBeforeChallenge(task *Task) error {
	task.Step = TASKWAITCHALLENGEBLOCK
	w.scanner.TaskFromWorkerCh <- task
	return nil
}

func (w *Worker) SolveProblem(task *Task) error {
	if task.Step != TASKGETCHALLENGEBLOCK {
		return InvalidStepError
	}
	// preimage: address || challenge hash
	addrBytes := task.coinbase.Bytes()
	preimage := append(addrBytes, task.lottery.ChallengHeaderHash[:]...)
	mimcHash, proof := w.zkpProver.Prove(preimage)
	if mimcHash == nil || proof == nil {
		return ZKPProofError
	}
	task.lottery.ZkpProof = proof
	task.lottery.MimcHash = mimcHash
	task.Step = TASKPROBLEMSOLVED
	return nil
}

func (w *Worker) SubmitTask(task *Task) error {
	return nil
}

func GetChallengeIndex(index [32]byte, interval uint64) uint64 {
	n := new(big.Int).SetBytes(index[:])
	module := new(big.Int).SetUint64(interval)
	return new(big.Int).Mod(n, module).Uint64()
}
