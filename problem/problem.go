package problem

import (
	"bytes"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
)

func CompileCircuit() frontend.CompiledConstraintSystem {
	var mimcCircuit Circuit
	r1cs, err := frontend.Compile(ecc.BN254, backend.GROTH16, &mimcCircuit)
	if err != nil {
		return nil
	}
	return r1cs
}

func SetupZKP(r1cs frontend.CompiledConstraintSystem) (groth16.ProvingKey, groth16.VerifyingKey) {
	pk, vk, err := groth16.Setup(r1cs)
	if err != nil {
		return nil, nil
	}
	return pk, vk
}

func NewProvingKey(b []byte) groth16.ProvingKey {
	k := groth16.NewProvingKey(ecc.BN254)
	buf := bytes.Buffer{}
	buf.Write(b)
	_, err := k.ReadFrom(&buf)
	if err != nil {
		return nil
	}
	return k
}

func NewVerifyingKey(b []byte) groth16.VerifyingKey {
	k := groth16.NewVerifyingKey(ecc.BN254)
	buf := bytes.Buffer{}
	buf.Write(b)
	_, err := k.ReadFrom(&buf)
	if err != nil {
		return nil
	}
	return k
}

func ZKPProve(r1cs frontend.CompiledConstraintSystem, pk groth16.ProvingKey, preimage []byte) ([]byte, []byte) {
	var c Circuit
	c.PreImage.Assign(preimage)
	hash := MimcHash(preimage)
	c.Hash.Assign(hash)
	proof, err := groth16.Prove(r1cs, pk, &c)
	if err != nil {
		return nil, nil
	}
	buf := bytes.Buffer{}
	proof.WriteTo(&buf)
	return hash, buf.Bytes()
}

func ZKPVerify(vk groth16.VerifyingKey, hash []byte, proof []byte) bool {
	p := groth16.NewProof(ecc.BN254)
	buf := bytes.Buffer{}
	buf.Write(proof)
	_, err := p.ReadFrom(&buf)
	if err != nil {
		return false
	}
	var c Circuit
	c.Hash.Assign(hash)
	err = groth16.Verify(p, vk, &c)
	if err != nil {
		return false
	}
	return true
}
