package problem

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
)

var (
	max256    = new(big.Int).Exp(big.NewInt(2), big.NewInt(256), big.NewInt(0))
	keccak256 = crypto.Keccak256
)

type Lottery struct {
	CoinbaseAddr       common.Address
	MinerAddr          common.Address //20 bytes
	ChallengHeaderHash [32]byte       //challenge block header Hash
	Index              [32]byte
	MimcHash           []byte //32 bytes
	ZkpProof           []byte //
	VrfProof           []byte
	Extra              []byte
}

func (l *Lottery) SetMinerAddr(addr common.Address) {
	l.MinerAddr = addr
}

func (l *Lottery) SetVrfProof(proof []byte) {
	l.VrfProof = proof
}

func (l *Lottery) SetZKPProof(proof []byte) {
	l.ZkpProof = proof
}

func (l *Lottery) SetExtra(extra []byte) {
	l.Extra = extra
}

func (l *Lottery) Serialize() {

}

func (l *Lottery) Deserialize() {

}

func (l *Lottery) Score() *big.Int {
	result := append(l.MinerAddr.Bytes(), l.MimcHash...)
	result = xor(keccak256(result), l.ChallengHeaderHash[:])
	return new(big.Int).SetBytes(result)
}

func IfPassDiff(score []byte, diff *big.Int) bool {
	target := new(big.Int).Div(max256, diff)
	if new(big.Int).SetBytes(score).Cmp(target) > 0 {
		return false
	} else {
		return true
	}
}

func xor(one, other []byte) (xor []byte) {
	if len(one) != len(other) {
		return nil
	}
	xor = make([]byte, len(one))
	for i := 0; i < len(one); i++ {
		xor[i] = one[i] ^ other[i]
	}
	return xor
}
