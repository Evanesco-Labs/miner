package problem

import (
	"bytes"
	"errors"
	"github.com/Evanesco-Labs/Miner/log"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"os"
)

var (
	ErrorInvalidR1CSPath = errors.New("r1cs path invalid")
	ErrorInvalidPkPath   = errors.New("proving key path invalid")
	ErrorInvalidVkPath   = errors.New("verifying key path invalid")
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
	mimchash := MimcHasher.Hash(preimage)
	c.Hash.Assign(mimchash)
	proof, err := groth16.Prove(r1cs, pk, &c)
	if err != nil {
		log.Error("groth16 error:", err.Error())
		return nil, nil
	}
	buf := bytes.Buffer{}
	proof.WriteTo(&buf)
	return mimchash, buf.Bytes()
}

func ZKPVerify(vk groth16.VerifyingKey, preimage []byte, hash []byte, proof []byte) bool {
	p := groth16.NewProof(ecc.BN254)
	buf := bytes.Buffer{}
	buf.Write(proof)
	_, err := p.ReadFrom(&buf)
	if err != nil {
		return false
	}
	var c Circuit
	c.Hash.Assign(hash)
	c.PreImage.Assign(preimage)
	err = groth16.Verify(p, vk, &c)
	if err != nil {
		return false
	}
	return true
}

type ProblemProver struct {
	r1cs frontend.CompiledConstraintSystem
	pk   groth16.ProvingKey
}

func (p *ProblemProver) Prove(preimage []byte) ([]byte, []byte) {
	return ZKPProve(p.r1cs, p.pk, preimage)
}

func NewProblemProver(pkPath string) (*ProblemProver, error) {
	r1cs := CompileCircuit()
	pkFile, err := os.OpenFile(pkPath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, ErrorInvalidPkPath
	}
	defer pkFile.Close()

	pk := groth16.NewProvingKey(ecc.BN254)
	_, err = pk.ReadFrom(pkFile)
	if err != nil {
		return nil, err
	}
	return &ProblemProver{
		r1cs: r1cs,
		pk:   pk,
	}, nil
}

type ProblemVerifier struct {
	vk groth16.VerifyingKey
}

func NewProblemVerifier(vkPath string) (*ProblemVerifier, error) {
	vkFile, err := os.OpenFile(vkPath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, ErrorInvalidVkPath
	}
	defer vkFile.Close()
	vk := groth16.NewVerifyingKey(ecc.BN254)
	_, err = vk.ReadFrom(vkFile)
	if err != nil {
		return nil, err
	}
	return &ProblemVerifier{
		vk: vk,
	}, nil
}

func (v *ProblemVerifier) Verify(preimage []byte, mimcHash []byte, proof []byte) bool {
	//hashExp := MimcHasher.Hash(preimage)
	//if !bytes.Equal(hashExp, mimcHash) {
	//	return false
	//}
	return ZKPVerify(v.vk, preimage, mimcHash, proof)
}
