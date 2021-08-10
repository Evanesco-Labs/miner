package main

import (
	"fmt"
	miner "github.com/Evanesco-Labs/Miner"
	"github.com/Evanesco-Labs/Miner/keypair"
	"github.com/Evanesco-Labs/Miner/log"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
)

var (
	testFlag = cli.BoolFlag{
		Name:  "test",
		Usage: "test mining",
	}
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
		testFlag,
		configFlag,
		urlFlag,
		pkFlag,
		minerKeyFlag,
		coinbaseFlag,
	},
	Action: StartMining,
}

//todo:load config file
func StartMining(ctx *cli.Context) {
	log.InitLog(0, os.Stdout, log.PATH)

	url := ctx.String("url")
	minerKeyPath := ctx.String("key")
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
	coinbaseStr := ctx.String("coinbase")
	//coinbase address is miner address by default
	if coinbaseStr == "" {
		coinbase = minerKey.Address
	} else {
		coinbase = common.HexToAddress(coinbaseStr)
	}

	pkPath := ctx.String("pk")

	config := miner.DefaultConfig()
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
