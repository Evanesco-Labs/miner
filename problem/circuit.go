package problem

import (
	"github.com/consensys/gnark-crypto/ecc"
	bn254 "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash/mimc"
	"hash"
	"sync"
)

const SEED = "Go Evanesco"

type Hasher struct {
	mu   sync.Mutex
	hash hash.Hash
}

var MimcHasher = Hasher{
	mu:   sync.Mutex{},
	hash: bn254.NewMiMC(SEED),
}

func (h *Hasher) Hash(m []byte) []byte {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.hash.Reset()
	h.hash.Write(m)
	return h.hash.Sum(nil)
}

type Circuit struct {
	PreImage frontend.Variable `gnark:",public"`
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
