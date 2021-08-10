package problem

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/Evanesco-Labs/Miner/log"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/stretchr/testify/assert"
	"os"
	"runtime"
	"testing"
	"time"
)

var (
	PkPath   = "./provekey.txt"
	VkPath   = "./verifykey.txt"
	R1csPath = "./r1cs.txt"
)

func TestNewProvingKey(t *testing.T) {
	r1cs := CompileCircuit()
	pk, vk := SetupZKP(r1cs)
	buf := bytes.Buffer{}
	pk.WriteTo(&buf)
	pkR := NewProvingKey(buf.Bytes())
	assert.Equal(t, false, pkR.IsDifferent(pk))

	buf.Reset()
	vk.WriteTo(&buf)
	vkR := NewVerifyingKey(buf.Bytes())
	assert.Equal(t, false, vkR.IsDifferent(vk))
}

func TestZKP(t *testing.T) {
	r1cs := CompileCircuit()
	pk, vk := SetupZKP(r1cs)

	msg := "test evanesco"
	hash := sha256.New()
	hash.Write([]byte(msg))
	preimage := hash.Sum(nil)

	mimcHash, proof := ZKPProve(r1cs, pk, preimage)

	result := ZKPVerify(vk,preimage, mimcHash, proof)
	assert.Equal(t, true, result)
}

func TestCircuit(t *testing.T) {
	log.InitLog(0, os.Stdout, log.PATH)
	runtime.GOMAXPROCS(1)
	var mimcCircuit Circuit

	r1cs, err := frontend.Compile(ecc.BN254, backend.GROTH16, &mimcCircuit)
	log.Debug("constraints:", r1cs.GetNbConstraints())
	if err != nil {
		t.Fatal(err)
	}

	startTime := time.Now()
	pk, vk, err := groth16.Setup(r1cs)
	if err != nil {
		t.Fatal(err)
	}
	log.Debug("setup time:", time.Now().Sub(startTime).String())

	{
		buf := bytes.Buffer{}
		n, err := pk.WriteTo(&buf)
		log.Debug("pk size:", n)

		buf = bytes.Buffer{}
		n, err = vk.WriteTo(&buf)
		log.Debug("vk size:", n)

		var witness Circuit

		msg := "test evanesco"
		hash1 := sha256.New()
		hash1.Write([]byte(msg))
		pre := hash1.Sum(nil)

		witness.PreImage.Assign(pre)

		hash := mimc.NewMiMC(SEED)
		var preimage fr.Element
		preimage.SetBytes(pre)
		pr := preimage.Bytes()
		hash.Write(pr[:])
		sum := hash.Sum(nil)
		witness.Hash.Assign(sum)

		start := time.Now()
		proof, err := groth16.Prove(r1cs, pk, &witness)
		if err != nil {
			t.Fatal(err)
		}
		duration := time.Now().Sub(start).String()
		log.Debug("prove time:", duration)

		buf = bytes.Buffer{}
		_, err = proof.WriteTo(&buf)
		if err != nil {
			log.Error(err)
		}

		var witnessV Circuit
		witnessV.PreImage.Assign(pre)
		witnessV.Hash.Assign(sum)
		start = time.Now()
		err = groth16.Verify(proof, vk, &witnessV)
		if err != nil {
			t.Fatal(err)
		}
		log.Debug("verify time:", time.Now().Sub(start).String())
	}
}

func TestNewProverVerifier(t *testing.T) {
	runtime.GOMAXPROCS(1)
	t1 := time.Now()
	prover, err := NewProblemProver(PkPath)
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
	verifier, err := NewProblemVerifier(VkPath)
	if err != nil {
		t.Fatal(err)
	}
	t4 := time.Now()
	result := verifier.Verify(preimage, mimcHash, proof)
	t5 := time.Now()
	fmt.Println("new prover:", t2.Sub(t1).String())
	fmt.Println("prove:", t3.Sub(t2).String())
	fmt.Println("new verifier:", t4.Sub(t3).String())
	fmt.Println("verify:", t5.Sub(t4).String())
	assert.Equal(t, true, result)
}

//TestSetupZKP checks setup randomness
func TestSetupZKP(t *testing.T) {
	r1cs := CompileCircuit()
	pk1, vk1 := SetupZKP(r1cs)

	pk2, vk2 := SetupZKP(r1cs)

	buf1 := new(bytes.Buffer)
	buf2 := new(bytes.Buffer)

	pk1.WriteTo(buf1)
	pk2.WriteTo(buf2)
	assert.Equal(t, false, bytes.Equal(buf1.Bytes(), buf2.Bytes()))

	buf1.Reset()
	buf2.Reset()
	vk1.WriteTo(buf1)
	vk2.WriteTo(buf2)
	assert.Equal(t, false, bytes.Equal(buf1.Bytes(), buf2.Bytes()))
}
