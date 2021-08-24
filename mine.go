package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	miner "github.com/ethereum/go-ethereum/zkpminer"
	"github.com/ethereum/go-ethereum/zkpminer/keypair"
	"gopkg.in/urfave/cli.v1"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

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
		passphraseFlag,
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
	runtime.GOMAXPROCS(1)
	config := miner.DefaultConfig()

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

	minerKey, err := GetKeyFromFile(ctx, minerKeyPath)
	if err != nil {
		utils.Fatalf("load miner key from file err %v", err)
	}

	minerKeyList := []keypair.Key{minerKey}

	var coinbase common.Address

	//coinbase address is miner address by default
	if coinbaseStr == "" {
		coinbase = minerKey.Address
	} else {
		coinbase = common.HexToAddress(coinbaseStr)
	}

	config.Customize(minerKeyList, coinbase, url, pkPath)

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
