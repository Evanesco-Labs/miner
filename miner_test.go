package miner

import (
	"github.com/Evanesco-Labs/Miner/keypair"
	"github.com/Evanesco-Labs/Miner/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/params"
	"os"
	"testing"
	"time"
)

var (
	testdb     = rawdb.NewMemoryDatabase()
	gspec      = &core.Genesis{Config: params.TestChainConfig}
	genesis    = gspec.MustCommit(testdb)
	blocks, _  = core.GenerateChain(params.TestChainConfig, genesis, ethash.NewFaker(), testdb, 100, nil)
	testURL    = "ws://127.0.0.1:8549"
	testPKPath = "./provekeytest.txt"
)

func DefaultTestConfig() Config {
	return Config{
		MinerList:        make([]keypair.Key, 0),
		MaxWorkerCnt:     10,
		MaxTaskCnt:       10,
		CoinbaseInterval: uint64(5),
		SubmitAdvance:    SUBMITADVANCE,
		CoinbaseAddr:     common.Address{},
		WsUrl:            "",
		RpcTimeout:       RPCTIMEOUT,
		PkPath:           "./provekey.txt",
	}
}

func TestMiner(t *testing.T) {
	log.InitLog(0, os.Stdout, log.PATH)

	minerList := make([]keypair.Key, 0)
	{
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

	config := DefaultTestConfig()
	config.Customize(minerList, coinbaseKey.Address, testURL, testPKPath)

	miner, err := NewMiner(config)
	if err != nil {
		t.Fatal(err)
	}
	log.Debug("coinbase address:", miner.scanner.CoinbaseAddr)

	//add another worker
	time.Sleep(time.Second * 20)

	_, sk := keypair.GenerateKeyPair()
	key, err := keypair.NewKey(sk.PrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	miner.NewWorker(key)

	select {}
}
