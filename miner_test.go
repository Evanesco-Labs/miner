package miner

import (
	"github.com/Evanesco-Labs/miner/keypair"
	"github.com/Evanesco-Labs/miner/log"
	"github.com/ethereum/go-ethereum/common"
	"os"
	"testing"
	"time"
)

var (
	testURL    = "ws://127.0.0.1:8549"
	testPKPath = "./provekeytest.txt"
)

func defaultTestConfig() Config {
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

	config := defaultTestConfig()
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
