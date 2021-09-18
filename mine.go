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
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

var AvisWsURL = []string{"ws://3.128.151.194:7778", "ws://18.118.180.198:7778", "ws://18.191.121.77:7778", "ws://18.191.17.148:7778", "ws://18.219.122.241:7778"}

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
		Value: "./QmNpJg4jDFE4LMNvZUzysZ2Ghvo4UJFcsjguYcx4dTfwKx",
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
	Url             []string `yaml:"url"`
	ZKPProveKeyPath string   `yaml:"zkp_prove_key"`
	MinerKey        string   `yaml:"miner_key"`
	CoinbaseAddress string   `yaml:"coinbase_address"`
}

func StartMining(ctx *cli.Context) {
	runtime.GOMAXPROCS(1)
	config := miner.DefaultConfig()

	urlList := make([]string, 0)
	if ctx.IsSet(urlFlag.Name) {
		urlList = append(urlList, ctx.String(urlFlag.Name))
	}

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
			if len(configYml.Url) != 0 {
				urlList = append(urlList, configYml.Url...)
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
	coinbase = common.HexToAddress(coinbaseStr)

	//random AvisWsURL order
	if len(urlList) == 0 {
		mixAvisUrl := MixSlice(AvisWsURL)
		urlList = append(urlList, mixAvisUrl...)
	}

	config.Customize(minerKeyList, coinbase, urlList, pkPath)

	_, err = miner.NewMiner(config)
	if err != nil {
		utils.Fatalf("Starting miner error: %v", err)
	}
	waitToExit()
}

func MixSlice(urlOrdered []string) []string {
	urlMix := make([]string, 0)
	l := len(urlOrdered)
	for i := 0; i < l; i++ {
		index := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(urlOrdered))
		urlMix = append(urlMix, urlOrdered[index])
		if index == len(urlOrdered)-1 {
			urlOrdered = urlOrdered[:index]
		} else {
			urlOrdered = append(urlOrdered[:index], urlOrdered[index+1:]...)
		}
	}
	return urlMix
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
