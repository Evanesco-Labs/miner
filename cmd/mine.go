package main

import (
	"fmt"
	miner "github.com/Evanesco-Labs/miner"
	"github.com/Evanesco-Labs/miner/keypair"
	"github.com/Evanesco-Labs/miner/log"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"gopkg.in/urfave/cli.v1"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

const BlockInterval = time.Second * 6

var (
	configFlag = cli.StringFlag{
		Name:  "config",
		Usage: "config file path",
		Value: "./config.yml",
	}
	urlFlag = cli.StringFlag{
		Name:  "url",
		Usage: "rpc url of Evanesco node",
	}
	pkFlag = cli.StringFlag{
		Name:  "pk",
		Usage: "provekey of ZKP problem path",
		Value: "./provekey.txt",
	}
	minerKeyFlag = cli.StringFlag{
		Name:  "key",
		Usage: "miner eva keyfile path",
		Value: "./keyfile.json",
	}
	coinbaseFlag = cli.StringFlag{
		Name:  "coinbase",
		Usage: "coinbase address to get mining rewards",
	}
)

var commandMine = cli.Command{
	Name:      "mine",
	Usage:     "start mining",
	ArgsUsage: "[ <keyfile> ]",
	Description: `
Start mining

If you want to test mining without Evanesco network, set the --test flag. 
`,
	Flags: []cli.Flag{
		configFlag,
		urlFlag,
		pkFlag,
		minerKeyFlag,
		coinbaseFlag,
	},
	Action: StartMining,
}

type ConfigYML struct {
	Url             string `yaml:"url"`
	ZKPProveKeyPath string `yaml:"zkp_prove_key"`
	MinerKey        string `yaml:"miner_key"`
	CoinbaseAddress string `yaml:"coinbase_address"`
}

func StartMining(ctx *cli.Context) {
	log.InitLog(log.InfoLog, os.Stdout, log.PATH)
	runtime.GOMAXPROCS(1)
	config := miner.DefaultTestConfig()

	url := ctx.String("url")
	minerKeyPath := ctx.String("key")
	coinbaseStr := ctx.String("coinbase")
	pkPath := ctx.String("pk")

	//load config from yaml file
	configPath := ctx.String("config")
	b, err := ioutil.ReadFile(configPath)
	var configYml ConfigYML
	if err == nil {
		err = yaml.Unmarshal(b, &configYml)
		if err == nil {
			if (!ctx.IsSet("url")) && configYml.Url != "" {
				url = configYml.Url
			}
			if (!ctx.IsSet("key")) && configYml.MinerKey != "" {
				minerKeyPath = configYml.MinerKey
			}
			if (!ctx.IsSet("coinbase")) && configYml.CoinbaseAddress != "" {
				coinbaseStr = configYml.CoinbaseAddress
			}
			if (!ctx.IsSet("pk")) && configYml.ZKPProveKeyPath != "" {
				pkPath = configYml.ZKPProveKeyPath
			}
		}
	}

	// Read key from file.
	keyJson, err := ioutil.ReadFile(minerKeyPath)
	if err != nil {
		utils.Fatalf("Failed to read the keyfile at '%s': %v", minerKeyPath, err)
	}

	// Decrypt key with passphrase.
	passphrase := getPassphrase(ctx, false)
	rawKey, err := keystore.DecryptKey(keyJson, passphrase)
	if err != nil {
		utils.Fatalf("Error decrypting key: %v", err)
	}

	minerKey, err := keypair.NewKey(rawKey.PrivateKey)
	if err != nil {
		utils.Fatalf("Error loading miner key : %v", err)
	}
	minerKeyList := []keypair.Key{minerKey}

	var coinbase common.Address

	//coinbase address is miner address by default
	if coinbaseStr == "" {
		coinbase = minerKey.Address
	} else {
		coinbase = common.HexToAddress(coinbaseStr)
	}

	headerCh := make(chan *types.Header)
	//start local test chain
	testdb := rawdb.NewMemoryDatabase()
	gspec := &core.Genesis{Config: params.TestChainConfig}
	genesis := gspec.MustCommit(testdb)
	hc, err := core.NewHeaderChain(testdb, params.TestChainConfig, ethash.NewFaker(), func() bool { return false })
	if err != nil {
		panic(err)
	}
	hc.SetGenesis(genesis.Header())
	parent := genesis

	go func() {
		ticker := time.Tick(BlockInterval)
		<-miner.FinishInitProver
		for {
			headerCh <- hc.CurrentHeader()
			<-ticker
			blocks, _ := core.GenerateChain(params.TestChainConfig, parent, ethash.NewFaker(), testdb, 1, nil)
			parent = blocks[0]

			headers := []*types.Header{blocks[0].Header()}
			_, err := hc.InsertHeaderChain(headers, time.Now())
			if err != nil {
				panic(err)
			}
		}
	}()

	getHeader := func(height miner.Height) (*types.Header, error) {
		return hc.GetHeaderByNumber(uint64(height)), nil
	}

	config.Customize(minerKeyList, coinbase, url, pkPath, getHeader, headerCh)

	_, err = miner.NewMiner(config)
	if err != nil {
		utils.Fatalf("Starting miner error: %v", err)
	}
	waitToExit()
}

func waitToExit() {
	exit := make(chan bool, 0)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		for sig := range sc {
			fmt.Printf("received exit signal:%v", sig.String())
			close(exit)
			break
		}
	}()
	<-exit
}
