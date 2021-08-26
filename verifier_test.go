package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/zkpminer/problem"
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
	"time"
)

var (
	PkPath = "../provekeytest.txt"
	VkPath = "../verifykeytest.txt"
)

func GenerateTestChain(cnt int) (*core.HeaderChain, error) {
	testdb := rawdb.NewMemoryDatabase()
	gspec := &core.Genesis{Config: params.TestChainConfig}
	genesis := gspec.MustCommit(testdb)
	hc, err := core.NewHeaderChain(testdb, params.TestChainConfig, ethash.NewFaker(), func() bool { return false })
	if err != nil {
		return nil, err
	}
	hc.SetGenesis(genesis.Header())
	blocks, _ := core.GenerateChain(params.TestChainConfig, genesis, ethash.NewFaker(), testdb, 100, nil)
	headerChain := make([]*types.Header, 0)
	for _, block := range blocks {
		headerChain = append(headerChain, block.Header())
	}

	_, err = hc.InsertHeaderChain(headerChain, time.Now())
	if err != nil {
		return nil, err
	}
	return hc, nil
}

func TestNewProverVerifier(t *testing.T) {
	hc, err := GenerateTestChain(100)
	if err != nil {
		panic(err)
	}
	runtime.GOMAXPROCS(1)
	t1 := time.Now()
	prover, err := problem.NewProblemProver(PkPath)
	t2 := time.Now()
	if err != nil {
		t.Fatal(err)
	}

	msg := "test evanesco"
	hash := sha256.New()
	hash.Write([]byte(msg))
	preimage := hash.Sum(nil)

	mimcHash, proof := prover.Prove(preimage)
	t3 := time.Now()
	getHeader := func(height uint64) (*types.Header, error) {
		header := hc.GetHeaderByNumber(height)
		if header == nil {
			return nil, errors.New("no header")
		}
		return header, nil
	}
	verifier, err := problem.NewProblemVerifier(VkPath, 100, 80, getHeader)
	if err != nil {
		t.Fatal(err)
	}
	t4 := time.Now()
	result := verifier.VerifyZKP(preimage, mimcHash, proof)
	t5 := time.Now()
	fmt.Println("new prover:", t2.Sub(t1).String())
	fmt.Println("prove:", t3.Sub(t2).String())
	fmt.Println("new verifier:", t4.Sub(t3).String())
	fmt.Println("verify:", t5.Sub(t4).String())
	assert.Equal(t, true, result)
}