package miner

import (
	"bytes"
	"errors"
	"github.com/Evanesco-Labs/Miner/keypair"
	"github.com/Evanesco-Labs/Miner/log"
	"github.com/Evanesco-Labs/Miner/problem"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"os"
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
	log.InitLog(0, os.Stdout, log.PATH)
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
			log.Debug("chain height:", height)
			chain[Height(height)] = block.Header()
		}
	}()

	getHeader := func(height Height) (*types.Header, error) {
		if header, ok := chain[height]; ok {
			return header, nil
		}
		return nil, errors.New("no such header")
	}

	minerList := make([]keypair.Key, 0)
	for i := 0; i < testWorkersCnt; i++ {
		_, sk := keypair.GenerateKeyPair()
		key, err := keypair.NewKey(sk.PrivateKey)
		if err != nil {
			t.Fatal(err)
		}
		minerList = append(minerList, key)
	}

	_, coinbaseSk := keypair.GenerateKeyPair()
	coinbaseKey, err := keypair.NewKey(coinbaseSk.PrivateKey)
	if err != nil {
		t.Fatal(err)
	}

	config := Config{
		MinerList:        minerList,
		MaxWorkerCnt:     int32(testWorkersCnt) + 1,
		MaxTaskCnt:       1,
		CoinbaseAddr:     coinbaseKey.Address,
		CoinbaseInterval: interval,
		SubmitAdvance:    uint64(1),
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
	log.Debug("coinbase address: ",miner.scanner.CoinbaseAddr)

	//add another worker
	time.Sleep(time.Second * 6)

	_, sk := keypair.GenerateKeyPair()
	key, err := keypair.NewKey(sk.PrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	miner.NewWorker(key)

	select {}
}
