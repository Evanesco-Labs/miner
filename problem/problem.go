package problem

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/Evanesco-Labs/miner/keypair"
	"github.com/Evanesco-Labs/miner/log"
	"github.com/Evanesco-Labs/miner/vrf"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"os"
)

type GetHeaderByNum = func(uint642 uint64) (*types.Header, error)

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

type Prover struct {
	r1cs frontend.CompiledConstraintSystem
	pk   groth16.ProvingKey
}

func (p *Prover) Prove(preimage []byte) ([]byte, []byte) {
	return ZKPProve(p.r1cs, p.pk, preimage)
}

func NewProblemProver(pkPath string) (*Prover, error) {
	log.Info("Compiling ZKP circuit")
	r1cs := CompileCircuit()
	pkFile, err := os.OpenFile(pkPath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, ErrorInvalidPkPath
	}
	defer pkFile.Close()
	log.Info("Loading ZKP prove key. This takes a few minutes")
	pk := groth16.NewProvingKey(ecc.BN254)
	_, err = pk.ReadFrom(pkFile)
	if err != nil {
		return nil, err
	}
	return &Prover{
		r1cs: r1cs,
		pk:   pk,
	}, nil
}

type Verifier struct {
	coinbaseInterval uint64
	submitAdvance    uint64
	vk               groth16.VerifyingKey
	getHeaderByNum   GetHeaderByNum
}

func NewProblemVerifier(vkPath string, interval, advance uint64, getHeaderByNum GetHeaderByNum) (*Verifier, error) {
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
	return &Verifier{
		coinbaseInterval: interval,
		submitAdvance:    advance,
		vk:               vk,
		getHeaderByNum:   getHeaderByNum,
	}, nil
}

func (v *Verifier) VerifyZKP(preimage []byte, mimcHash []byte, proof []byte) bool {
	return ZKPVerify(v.vk, preimage, mimcHash, proof)
}

//todo: add additional check if the lottery miner pledged
func (v *Verifier) VerifyLottery(lottery *types.Lottery, sigBytes []byte, lastCoinbaseHeader *types.Header) bool {
	msg, err := json.Marshal(lottery)
	if err != nil {
		return false
	}

	msgHash :=  crypto.Keccak256(msg)
	ecdsaPK, err := crypto.SigToPub(msgHash, sigBytes)
	if err != nil {
		return false
	}
	pk, err := keypair.NewPublicKey(ecdsaPK)
	if err != nil {
		return false
	}

	if crypto.PubkeyToAddress(*ecdsaPK) != lottery.MinerAddr {
		return false
	}

	lastCoinbaseHash := lastCoinbaseHeader.Hash()
	index, err := vrf.ProofToHash(pk, lastCoinbaseHash[:], lottery.VrfProof)
	if err != nil {
		return false
	}

	if index != lottery.Index {
		return false
	}

	challengeHeight := lastCoinbaseHeader.Number.Uint64() + GetChallengeIndex(index, uint64(v.coinbaseInterval)-uint64(v.submitAdvance))
	challengeHeader, err := v.getHeaderByNum(challengeHeight)
	if err != nil || challengeHeader == nil {
		return false
	}

	if challengeHeader.Hash() != lottery.ChallengeHeaderHash {
		return false
	}

	addrBytes := lottery.MinerAddr.Bytes()
	preimage := append(addrBytes, lottery.ChallengeHeaderHash[:]...)
	preimage = crypto.Keccak256(preimage)
	return v.VerifyZKP(preimage, lottery.MimcHash, lottery.VrfProof)
}

func GetChallengeIndex(index [32]byte, interval uint64) uint64 {
	n := new(big.Int).SetBytes(index[:])
	module := new(big.Int).SetUint64(interval)
	return new(big.Int).Mod(n, module).Uint64()
}
