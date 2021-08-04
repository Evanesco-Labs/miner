package problem

import (
	"github.com/consensys/gnark-crypto/ecc"
	bn254 "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash/mimc"
)

const SEED = "Go Evanesco"

var hash = bn254.NewMiMC(SEED)

// MimcHash output 32 bytes
func MimcHash(m []byte) []byte {
	hash.Reset()
	hash.Write(m)
	return hash.Sum(nil)
}

type Circuit struct {
	PreImage frontend.Variable
	Hash     frontend.Variable `gnark:",public"`
}

// Define declares the circuit's constraints
// Hash = mimc(PreImage)
func (circuit *Circuit) Define(curveID ecc.ID, cs *frontend.ConstraintSystem) error {
	// hash function
	mimcIns, _ := mimc.NewMiMC(SEED, curveID)

	// specify constraints
	// mimc(preImage) == hash
	cs.AssertIsEqual(circuit.Hash, mimcIns.Hash(cs, circuit.PreImage))
	return nil
}
