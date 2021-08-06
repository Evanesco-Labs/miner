package miner

import (
	"errors"
	"fmt"
	"github.com/Evanesco-Labs/Miner/problem"
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
	CoinbaseInterval   Height
	LastCoinbaseHeight Height
	taskWait           map[Height][]*Task //single concurrent
	headerCh           chan *types.Header
	inboundTaskCh      chan *Task //channel to receive tasks from worker
	outboundTaskCh     chan *Task //channel to send challenged task to miner
	getHeader          func(height Height) (*types.Header, error)
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

func (s *Scanner) close() {
	close(s.exitCh)
}

func (s *Scanner) Loop() {
	for {
		select {
		case <-s.exitCh:
			return
		case header := <-s.headerCh:
			fmt.Println("best score: ", s.BestScore)
			height := Height(header.Number.Uint64())
			index := height - s.LastCoinbaseHeight
			fmt.Println("index: ", index)

			s.LastBlockHeight = height
			if s.IfCoinBase(header) {
				task := s.NewTask(header)
				s.taskWait = make(map[Height][]*Task)
				s.outboundTaskCh <- &task
				s.LastCoinbaseHeight = height
				s.BestScore = zero
			}

			if taskList, ok := s.taskWait[index]; ok {
				//add challengeHeader and send tasks to miner task channel
				for _, task := range taskList {
					if task.Step != TASKWAITCHALLENGEBLOCK {
						log.Error(InvalidTaskStepErr.Error())
						continue
					}
					task.SetHeader(header)
					s.outboundTaskCh <- task
				}
				delete(s.taskWait, index)
				continue
			}

		case task := <-s.inboundTaskCh:
			if task.Step == TASKWAITCHALLENGEBLOCK {
				if taskList, ok := s.taskWait[task.challengeIndex]; ok {
					taskList = append(taskList, task)
					s.taskWait[task.challengeIndex] = taskList
					continue
				}
				taskList := []*Task{task}
				s.taskWait[task.challengeIndex] = taskList
				continue
			}
			if task.Step == TASKPROBLEMSOLVED {
				taskScore := task.lottery.Score()
				fmt.Println("get solved score: ", taskScore)
				if taskScore.Cmp(s.BestScore) != 1 {
					fmt.Println("less than best", s.BestScore)
					continue
				}
				s.BestScore = taskScore
				err := s.Submit(task)
				if err != nil {
					// put this tasks to a list and wait for connection success
					log.Error(err.Error())
				}
				task.Step = TASKSUBMITTED
				s.outboundTaskCh <- task
			}
		}
	}
}

//todo: here only for testing, implement this
func (s *Scanner) IfCoinBase(h *types.Header) bool {
	return new(big.Int).Mod(h.Number, new(big.Int).SetUint64(uint64(s.CoinbaseInterval))).Cmp(zero) == 0
}

//todo: GetHeader get challengeHeader from remote or local node, returns nil if not yet this height.
func (s *Scanner) GetHeader(height Height) (*types.Header, error) {
	return s.getHeader(height)
}

//TODO: check if the best before submit
func (s *Scanner) Submit(task *Task) error {
	// Submit check if the lottery has the best score
	fmt.Println("submit lottery", task.coinbase, task.lottery.Score().String())
	return nil
}
