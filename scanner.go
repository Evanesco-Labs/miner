package miner

import (
	"context"
	"errors"
	"fmt"
	"github.com/Evanesco-Labs/miner/log"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
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

type Explorer interface {
	GetHeaderChan() chan *types.Header
	GetHeaderByNum(num uint64) *types.Header
}

type RpcExplorer struct {
	Client     *rpc.Client
	Sub        ethereum.Subscription
	headerCh   chan *types.Header
	rpcTimeOut time.Duration
	wsUrl      string
}

func (r *RpcExplorer) GetHeaderChan() chan *types.Header {
	return r.headerCh
}

func (r *RpcExplorer) GetHeaderByNum(num uint64) *types.Header {
	ctx, cancel := context.WithTimeout(context.Background(), r.rpcTimeOut)
	defer cancel()
	var head *types.Header
	err := r.Client.CallContext(ctx, &head, "eth_getBlockByNumber", toBlockNumArg(new(big.Int).SetUint64(num)), false)
	if err != nil {
		log.Error(err)
		return nil
	}
	return head
}

type LocalExplorer struct {
	Backend
	headerCh chan *types.Header
}

func (l *LocalExplorer) GetHeaderChan() chan *types.Header {
	return l.headerCh
}

func (l *LocalExplorer) GetHeaderByNum(num uint64) *types.Header {
	return l.BlockChain().GetHeaderByNumber(num)
}

type Scanner struct {
	mu                 sync.RWMutex
	CoinbaseAddr       common.Address
	BestScore          *big.Int
	LastBlockHeight    Height
	CoinbaseInterval   Height
	LastCoinbaseHeight Height
	explorer           Explorer
	taskWait           map[Height][]*Task //single concurrent
	inboundTaskCh      chan *Task         //channel to receive tasks from worker
	outboundTaskCh     chan *Task         //channel to send challenged task to miner
	exitCh             chan struct{}
}

func (s *Scanner) NewTask(h *types.Header) Task {
	return Task{
		Step:             TASKSTART,
		CoinbaseAddr:     s.CoinbaseAddr,
		lastCoinBaseHash: h.Hash(),
		challengeIndex:   Height(uint64(0)),
		lottery: &types.Lottery{
			CoinbaseAddr: s.CoinbaseAddr,
		},
	}
}

func (s *Scanner) close() {
	defer func() {
		if recover() != nil {
		}
	}()
	close(s.exitCh)
}

func (s *Scanner) Loop() {
	headerCh := s.explorer.GetHeaderChan()
	for {
		select {
		case <-s.exitCh:
			return
		case header := <-headerCh:
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
	header := s.explorer.GetHeaderByNum(uint64(height))
	if header == nil {
		return header, fmt.Errorf("get header failed, height: %v", height)
	}
	return header, nil
}

func (s *Scanner) Submit(task *Task) error {
	// Submit check if the lottery has the best score
	log.Info("submit work",
		"\nminer address:", task.minerAddr,
		"\ncoinbase address:", task.CoinbaseAddr,
		"\nscore:", task.lottery.Score().String(),
	)

	if localExp, ok := s.explorer.(*LocalExplorer); ok {
		err := localExp.EventMux().Post(core.NewSolvedLotteryEvent{Lot: types.LotterySubmit{
			Lottery:   *task.lottery,
			Signature: task.signature,
		},
		})
		if err != nil {
			log.Error("submit with local explorer", err)
		}
		return nil
	}
	if rpcExp, ok := s.explorer.(*RpcExplorer); ok {
		ctx, cancel := context.WithTimeout(context.Background(), RPCTIMEOUT)
		defer cancel()
		submit := types.LotterySubmit{
			Lottery:   *task.lottery,
			Signature: task.signature,
		}

		err := rpcExp.Client.CallContext(ctx, nil, "eth_lotterySubmit", submit)
		if err != nil {
			log.Error("rpc submit lottery error:", err)
		}
		return err
	}

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
