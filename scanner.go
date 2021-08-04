package miner

import (
	"errors"
	"github.com/Evanesco-Labs/Miner/problem"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
	"sync"
)

var zero = new(big.Int).SetInt64(int64(0))

var (
	InvalidTaskStepErr = errors.New("not ")
)

type Height uint64

type Scanner struct {
	mu                 sync.RWMutex
	BestScore          *big.Int
	LastBlockHeight    Height
	LastCoinbaseHeight Height
	taskWait           map[Height][]*Task //single concurrent
	headerCh           chan *types.Header
	TaskToWorkerCh     map[common.Address]chan *Task //channel to send task to worker
	TaskFromWorkerCh   chan *Task                    //channel to receive tasks from worker
	minerTaskCh        chan *Task                    //channel to send challenged task to miner
	startCh            chan struct{}
	stopCh             chan struct{}
	exitCh             chan struct{}
}

func (s *Scanner) NewTask(h *types.Header) Task {
	return Task{
		Step:             TASKSTART,
		lastCoinBaseHash: h.Hash(),
		challengeIndex:   Height(uint64(0)),
		lottery:          &problem.Lottery{},
	}
}

func (s *Scanner) Loop() {
	for {
		select {
		case <-s.startCh:
		case <-s.stopCh:
		case <-s.exitCh:
		case header := <-s.headerCh:
			height := Height(header.Number.Uint64())
			index := height - s.LastBlockHeight
			s.LastBlockHeight = height
			if IfCoinBase(header) {
				task := s.NewTask(header)
				s.minerTaskCh <- &task
				s.LastCoinbaseHeight = height
				s.BestScore = zero
				continue
			}

			if taskList, ok := s.taskWait[index]; ok {
				//add challengeHeader and send tasks to miner task channel
				for _, task := range taskList {
					if task.Step != TASKWAITCHALLENGEBLOCK {
						log.Error(InvalidTaskStepErr.Error())
						continue
					}
					task.SetHeader(header)
					s.minerTaskCh <- task
				}
				delete(s.taskWait, index)
				continue
			}

		case task := <-s.TaskFromWorkerCh:
			if taskList, ok := s.taskWait[task.challengeIndex]; ok {
				taskList = append(taskList, task)
				s.taskWait[task.challengeIndex] = taskList
				continue
			}
			taskList := []*Task{task}
			s.taskWait[task.challengeIndex] = taskList
		}
	}
}

func IfCoinBase(h *types.Header) bool {
	return false
}

//todo: GetHeader get challengeHeader from remote or local node, returns nil if not yet this height.
func (s *Scanner) GetHeader(height Height) (*types.Header, error) {
	return nil, nil
}

// todo: check if the best before submit
func (s *Scanner) Submit(task *Task) error {
	// Submit check if the lottery has the best score
	return nil
}
