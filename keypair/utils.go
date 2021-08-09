package keypair

import (
	"crypto/sha512"
	"encoding/binary"
	"math/big"
)

var one = big.NewInt(1)

func HashToCurve(m []byte) (x, y *big.Int) {
	h := sha512.New()
	var i uint32

	byteLen := (Params.BitSize + 7) >> 3
	for x == nil && i < 100 {
		h.Reset()
		binary.Write(h, binary.BigEndian, i)
		h.Write(m)
		r := []byte{2} // Set point encoding to "compressed", y=0.
		r = h.Sum(r)
		x, y = Unmarshal(Curve, r[:byteLen+1])
		i++
	}
	return
}

// HashToField hashes to an integer [1,N-1]
func HashToField(m []byte) *big.Int {
	// NIST SP 800-90A § A.5.1: Simple discard method.
	byteLen := (Params.BitSize + 7) >> 3
	h := sha512.New()
	for i := uint32(0); ; i++ {
		// TODO: Use a NIST specified DRBG.
		h.Reset()
		binary.Write(h, binary.BigEndian, i)
		h.Write(m)
		b := h.Sum(nil)
		k := new(big.Int).SetBytes(b[:byteLen])
		if k.Cmp(new(big.Int).Sub(Params.N, one)) == -1 {
			return k.Add(k, one)
		}
	}
}
