package problem

import (
	"bytes"
	"crypto/sha256"
	"github.com/Evanesco-Labs/Miner/log"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
	"time"
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

	result := ZKPVerify(vk, mimcHash, proof)
	assert.Equal(t, true, result)
}

func TestCircuit(t *testing.T) {
	runtime.GOMAXPROCS(1)
	var mimcCircuit Circuit

	r1cs, err := frontend.Compile(ecc.BN254, backend.GROTH16, &mimcCircuit)
	log.Debug("constraints: ", r1cs.GetNbConstraints())
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
		log.Debug("pk size: ", n)

		buf = bytes.Buffer{}
		n, err = vk.WriteTo(&buf)
		log.Debug("vk size: ", n)

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
		log.Debug("prove time: ", duration)

		buf = bytes.Buffer{}
		_, err = proof.WriteTo(&buf)
		if err != nil {
			log.Error(err)
		}

		start = time.Now()
		err = groth16.Verify(proof, vk, &witness)
		if err != nil {
			t.Fatal(err)
		}
		log.Debug("verify time: ", time.Now().Sub(start).String())
	}
}

func TestNewProblemProver(t *testing.T) {
	r1cs := CompileCircuit()
	pk, vk := SetupZKP(r1cs)

	r1csBuf := new(bytes.Buffer)
	r1cs.WriteTo(r1csBuf)

	pkBuf := new(bytes.Buffer)
	pk.WriteTo(pkBuf)

	prover, err := NewProblemProver(r1csBuf, pkBuf)
	if err != nil {
		t.Fatal(err)
	}

	msg := "test evanesco"
	hash := sha256.New()
	hash.Write([]byte(msg))
	preimage := hash.Sum(nil)

	mimcHash, proof := prover.Prove(preimage)

	assert.Equal(t, true, ZKPVerify(vk, mimcHash, proof))
}

func TestNewProblemVerifier(t *testing.T) {
	r1cs := CompileCircuit()
	pk, vk := SetupZKP(r1cs)

	vkBuf := new(bytes.Buffer)
	vk.WriteTo(vkBuf)

	verifier, err := NewProblemVerifier(vkBuf)
	if err != nil {
		t.Fatal(err)
	}

	msg := "test evanesco"
	hash := sha256.New()
	hash.Write([]byte(msg))
	preimage := hash.Sum(nil)

	hashMimc, proof := ZKPProve(r1cs, pk, preimage)

	assert.Equal(t, true, verifier.Verify(preimage, hashMimc, proof))
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
