package problem

import (
	"bytes"
	"crypto/sha256"
	"fmt"
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
	fmt.Println("constraints: ", r1cs.GetNbConstraints())
	if err != nil {
		t.Fatal(err)
	}

	startTime := time.Now()
	pk, vk, err := groth16.Setup(r1cs)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("setup time:", time.Now().Sub(startTime).String())

	{
		buf := bytes.Buffer{}
		n, err := pk.WriteTo(&buf)
		fmt.Println("pk size: ", n)

		buf = bytes.Buffer{}
		n, err = vk.WriteTo(&buf)
		fmt.Println("vk size: ", n)

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
		fmt.Println("prove time: ", duration)

		buf = bytes.Buffer{}
		_, err = proof.WriteTo(&buf)
		if err != nil {
			fmt.Println(err)
		}

		start = time.Now()
		err = groth16.Verify(proof, vk, &witness)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("verify time: ", time.Now().Sub(start).String())
	}
}
