package miner

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Evanesco-Labs/Miner/keypair"
	"github.com/Evanesco-Labs/Miner/problem"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"testing"
	"time"
)

var (
	testdb    = rawdb.NewMemoryDatabase()
	gspec     = &core.Genesis{Config: params.TestChainConfig}
	genesis   = gspec.MustCommit(testdb)
	blocks, _ = core.GenerateChain(params.TestChainConfig, genesis, ethash.NewFaker(), testdb, 100, nil)
)

func TestMiner(t *testing.T) {
	//test configs
	testWorkersCnt := 1
	interval := uint64(5)

	chain := make(map[Height]*types.Header)
	headerCh := make(chan *types.Header)

	go func() {
		ticker := time.Tick(time.Second)
		for _, block := range blocks {
			<-ticker
			headerCh <- block.Header()
			height := block.NumberU64()
			fmt.Println("chain height:", height)
			chain[Height(height)] = block.Header()
		}
	}()

	getHeader := func(height Height) (*types.Header, error) {
		if header, ok := chain[height]; ok {
			return header, nil
		}
		return nil, errors.New("no such header")
	}


	coinbaseList := make([]keypair.Key, 0)
	for i := 0; i < testWorkersCnt; i++ {
		_, sk := keypair.GenerateKeyPair()
		key, err := keypair.NewKey(sk.PrivateKey)
		if err != nil {
			t.Fatal(err)
		}
		coinbaseList = append(coinbaseList, key)
	}

	config := Config{
		CoinbaseList:     coinbaseList,
		MaxWorkerCnt:     int32(testWorkersCnt) + 1,
		MaxTaskCnt:       1,
		CoinbaseInterval: interval,
		SubmitAdvance: uint64(1),
		HeaderCh:         headerCh,
		getHeader:        getHeader,
	}

	r1cs := problem.CompileCircuit()
	pk, _ := problem.SetupZKP(r1cs)

	r1csBuf := new(bytes.Buffer)
	r1cs.WriteTo(r1csBuf)

	pkBuf := new(bytes.Buffer)
	pk.WriteTo(pkBuf)

	miner, err := NewMiner(config, r1csBuf, pkBuf)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(len(miner.workers))

	time.Sleep(time.Second * 6)
	//add another worker

	_, sk := keypair.GenerateKeyPair()
	key, err := keypair.NewKey(sk.PrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	miner.NewWorker(key)

	select {}
}
