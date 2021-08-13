package miner

import (
	"context"
	"errors"
	"github.com/Evanesco-Labs/miner/log"
	"github.com/Evanesco-Labs/miner/problem"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"math/big"
	"sync"
	"time"
)

var zero = new(big.Int).SetInt64(int64(0))

var (
	InvalidTaskStepErr = errors.New("not ")
)

type Height uint64

type Scanner struct {
	mu                 sync.RWMutex
	CoinbaseAddr       common.Address
	BestScore          *big.Int
	LastBlockHeight    Height
	CoinbaseInterval   Height
	LastCoinbaseHeight Height
	evaClient          *rpc.Client
	sub                ethereum.Subscription
	taskWait           map[Height][]*Task //single concurrent
	headerCh           chan *types.Header
	inboundTaskCh      chan *Task //channel to receive tasks from worker
	outboundTaskCh     chan *Task //channel to send challenged task to miner
	rpcTimeout         time.Duration
	wsUrl              string
	exitCh             chan struct{}
}

func (s *Scanner) NewTask(h *types.Header) Task {
	return Task{
		Step:             TASKSTART,
		CoinbaseAddr:     s.CoinbaseAddr,
		lastCoinBaseHash: h.Hash(),
		challengeIndex:   Height(uint64(0)),
		lottery: &problem.Lottery{
			CoinbaseAddr: s.CoinbaseAddr,
		},
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
		case err := <-s.sub.Err():
			//todo: improve robustness
			panic(err)
			return
		case header := <-s.headerCh:
			log.Debug("best score:", s.BestScore)
			height := Height(header.Number.Uint64())
			//index := height - s.LastCoinbaseHeight
			index := Height(new(big.Int).Mod(header.Number, new(big.Int).SetUint64(uint64(s.CoinbaseInterval))).Uint64())
			log.Info("chain height:", height, " index:", index)

			s.LastBlockHeight = height
			if s.IfCoinBase(header) {
				log.Info("start new mining epoch")
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
				log.Debug("get solved score:", taskScore)
				if taskScore.Cmp(s.BestScore) != 1 {
					log.Debug("less than best", s.BestScore)
					continue
				}
				s.BestScore = taskScore
				//todo:abort lottery if exceed deadline
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

func (s *Scanner) IfCoinBase(h *types.Header) bool {
	return new(big.Int).Mod(h.Number, new(big.Int).SetUint64(uint64(s.CoinbaseInterval))).Cmp(zero) == 0
}

//todo: improve robustness, add some retries
func (s *Scanner) GetHeader(height Height) (*types.Header, error) {
	num := new(big.Int).SetUint64(uint64(height))
	ctx, cancel := context.WithTimeout(context.Background(), s.rpcTimeout)
	defer cancel()
	var head *types.Header
	err := s.evaClient.CallContext(ctx, &head, "eth_getBlockByNumber", toBlockNumArg(num), false)
	if err == nil && head == nil {
		err = ethereum.NotFound
	}
	return head, err
}

func (s *Scanner) Submit(task *Task) error {
	// Submit check if the lottery has the best score
	log.Info("submit work\n",
		"\nminer:", task.minerAddr,
		"\ncoinbase address:", task.CoinbaseAddr,
		"\nscore:", task.lottery.Score().String(),
	)

	//todo: rpc call to submit work
	//ctx, cancel := context.WithTimeout(context.Background(), s.rpcTimeout)
	//defer cancel()
	//s.evaClient.CallContext(ctx)
	return nil
}

func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	pending := big.NewInt(-1)
	if number.Cmp(pending) == 0 {
		return "pending"
	}
	return hexutil.EncodeBig(number)
}
