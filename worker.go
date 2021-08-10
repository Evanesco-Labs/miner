package miner

import (
	"errors"
	"github.com/Evanesco-Labs/Miner/keypair"
	"github.com/Evanesco-Labs/Miner/log"
	"github.com/Evanesco-Labs/Miner/problem"
	"github.com/Evanesco-Labs/Miner/vrf"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"sync"
	"sync/atomic"
)

var (
	InvalidStepError = errors.New("invalid task step")
	ZKPProofError    = errors.New("zkp prove error")
)

type Worker struct {
	mu               sync.RWMutex
	running          int32
	MaxTaskCnt       int32
	CoinbaseAddr     common.Address
	minerAddr        common.Address
	pk               *keypair.PublicKey
	sk               *keypair.PrivateKey
	workingTaskCnt   int32
	coinbaseInterval Height
	submitAdvance    Height
	inboundTaskCh    chan *Task //channel to get task from miner
	scanner          *Scanner
	zkpProver        *problem.ProblemProver
	exitCh           chan struct{}
}

func (w *Worker) Loop() {
	for {
		select {
		case <-w.exitCh:
			atomic.StoreInt32(&w.running, 0)
			return
		case task := <-w.inboundTaskCh:
			if !w.isRunning() {
				continue
			}
			if task.Step == TASKSTART {
				go func() {
					atomic.AddInt32(&w.workingTaskCnt, int32(1))
					defer atomic.AddInt32(&w.workingTaskCnt, int32(-1))
					err := w.HandleStartTask(task)
					if err != nil {
						log.Error(err.Error())
					}
				}()
				continue
			}
			if task.Step == TASKGETCHALLENGEBLOCK {
				go func() {
					atomic.AddInt32(&w.workingTaskCnt, int32(1))
					defer atomic.AddInt32(&w.workingTaskCnt, int32(-1))
					err := w.HandleChallengedTask(task)
					if err != nil {
						log.Error(err.Error())
					}
				}()
				continue
			}
		}
	}
}

func (w *Worker) isRunning() bool {
	return atomic.LoadInt32(&w.running) == 1
}

func (w *Worker) start() {
	atomic.StoreInt32(&w.running, 1)
}

func (w *Worker) stop() {
	atomic.StoreInt32(&w.running, 0)
}

func (w *Worker) close() {
	atomic.StoreInt32(&w.running, 0)
	close(w.exitCh)
}

func (w *Worker) HandleStartTask(task *Task) error {
	log.Debug("handle start task")
	task.minerAddr = w.minerAddr
	task.lottery.SetMinerAddr(w.minerAddr)
	index, proof := vrf.Evaluate(w.sk, task.lastCoinBaseHash[:])
	task.challengeIndex = Height(GetChallengeIndex(index, uint64(w.coinbaseInterval)-uint64(w.submitAdvance)))

	task.lottery.VrfProof = proof
	task.lottery.Index = index
	task.Step = TASKWAITCHALLENGEBLOCK

	log.Debug("challenge height:", w.scanner.LastCoinbaseHeight+task.challengeIndex)
	// request if this block already exit
	if w.scanner.LastBlockHeight+task.challengeIndex <= w.scanner.LastBlockHeight {
		return w.HandleTaskAfterChallenge(task)
	}

	// waiting for challenge block exist
	return w.HandlerTaskBeforeChallenge(task)
}

func (w *Worker) HandleChallengedTask(task *Task) error {
	log.Debug("handler challenged task")
	// start zkp proof
	err := w.SolveProblem(task)
	if err != nil {
		log.Error(err)
		return err
	}

	//give it to miner to submit
	w.scanner.inboundTaskCh <- task
	return nil
}

func (w *Worker) HandleTaskAfterChallenge(task *Task) error {
	log.Debug("handler task after challenge")
	header, err := w.scanner.GetHeader(w.scanner.LastBlockHeight + task.challengeIndex)
	if err != nil {
		return err
	}
	task.SetHeader(header)
	return w.HandleChallengedTask(task)
}

func (w *Worker) HandlerTaskBeforeChallenge(task *Task) error {
	log.Debug("handle task before challenge, index ", task.challengeIndex)
	task.Step = TASKWAITCHALLENGEBLOCK
	w.scanner.inboundTaskCh <- task
	return nil
}

func (w *Worker) SolveProblem(task *Task) error {
	if task.Step != TASKGETCHALLENGEBLOCK {
		return InvalidStepError
	}
	// preimage: keccak(address || challenge hash)
	addrBytes := task.minerAddr.Bytes()
	preimage := append(addrBytes, task.lottery.ChallengHeaderHash[:]...)
	preimage = crypto.Keccak256(preimage)
	mimcHash, proof := w.zkpProver.Prove(preimage)
	if mimcHash == nil || proof == nil {
		return ZKPProofError
	}
	task.lottery.ZkpProof = proof
	task.lottery.MimcHash = mimcHash
	task.Step = TASKPROBLEMSOLVED
	return nil
}

func GetChallengeIndex(index [32]byte, interval uint64) uint64 {
	n := new(big.Int).SetBytes(index[:])
	module := new(big.Int).SetUint64(interval)
	return new(big.Int).Mod(n, module).Uint64()
}
