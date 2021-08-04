package vrf

import (
	"bytes"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"github.com/Evanesco-Labs/Miner/keypair"
	"math/big"
)

var (
	ErrInvalidVRF = errors.New("invalid VRF proof")
)

// Evaluate returns the verifiable unpredictable function evaluated at m
func Evaluate(k *keypair.PrivateKey, m []byte) (index [32]byte, proof []byte) {
	nilIndex := [32]byte{}
	// Prover chooses r <-- [1,N-1]
	r, _, _, err := elliptic.GenerateKey(keypair.Curve, rand.Reader)
	if err != nil {
		return nilIndex, nil
	}
	ri := new(big.Int).SetBytes(r)

	// H = HashToCurve(m)
	Hx, Hy := keypair.HashToCurve(m)
	// VRF_k(m) = [k]H
	sHx, sHy := keypair.Curve.ScalarMult(Hx, Hy, k.D.Bytes())
	vrf := elliptic.Marshal(keypair.Curve, sHx, sHy) // 65 bytes.

	// G is the base point
	// s = HashToField(G, H, [k]G, VRF, [r]G, [r]H)
	rGx, rGy := keypair.Curve.ScalarBaseMult(r)
	rHx, rHy := keypair.Curve.ScalarMult(Hx, Hy, r)
	var b bytes.Buffer
	b.Write(elliptic.Marshal(keypair.Curve, keypair.Params.Gx, keypair.Params.Gy))
	b.Write(elliptic.Marshal(keypair.Curve, Hx, Hy))
	b.Write(elliptic.Marshal(keypair.Curve, k.PublicKey.X, k.PublicKey.Y))
	b.Write(vrf)
	b.Write(elliptic.Marshal(keypair.Curve, rGx, rGy))
	b.Write(elliptic.Marshal(keypair.Curve, rHx, rHy))
	s := keypair.HashToField(b.Bytes())

	// t = r−s*k mod N
	t := new(big.Int).Sub(ri, new(big.Int).Mul(s, k.D))
	t.Mod(t, keypair.Params.N)

	// Index = H(vrf)
	index = sha256.Sum256(vrf)

	// Write s, t, and vrf to a proof blob. Also write leading zeros before s and t
	// if needed.
	var buf bytes.Buffer
	buf.Write(make([]byte, 32-len(s.Bytes())))
	buf.Write(s.Bytes())
	buf.Write(make([]byte, 32-len(t.Bytes())))
	buf.Write(t.Bytes())
	buf.Write(vrf)

	p := buf.Bytes()

	//self verify
	hash, err := ProofToHash(k.Public(), m, p)
	if err != nil {
		return nilIndex, nil
	}
	if hash != index {
		return nilIndex, nil
	}

	return index, p
}

// ProofToHash asserts that proof is correct for m and outputs index.
func ProofToHash(pk *keypair.PublicKey, m, proof []byte) (index [32]byte, err error) {
	nilIndex := [32]byte{}

	if got, want := len(proof), 64+65; got != want {
		return nilIndex, ErrInvalidVRF
	}

	s := proof[0:32]
	t := proof[32:64]
	vrf := proof[64 : 64+65]

	uHx, uHy := elliptic.Unmarshal(keypair.Curve, vrf)
	if uHx == nil {
		return nilIndex, ErrInvalidVRF
	}

	tGx, tGy := keypair.Curve.ScalarBaseMult(t)
	ksGx, ksGy := keypair.Curve.ScalarMult(pk.X, pk.Y, s)
	tksGx, tksGy := keypair.Curve.Add(tGx, tGy, ksGx, ksGy)

	Hx, Hy := keypair.HashToCurve(m)
	tHx, tHy := keypair.Curve.ScalarMult(Hx, Hy, t)
	sHx, sHy := keypair.Curve.ScalarMult(uHx, uHy, s)
	tksHx, tksHy := keypair.Curve.Add(tHx, tHy, sHx, sHy)

	var b bytes.Buffer
	b.Write(elliptic.Marshal(keypair.Curve, keypair.Params.Gx, keypair.Params.Gy))
	b.Write(elliptic.Marshal(keypair.Curve, Hx, Hy))
	b.Write(elliptic.Marshal(keypair.Curve, pk.X, pk.Y))
	b.Write(vrf)
	b.Write(elliptic.Marshal(keypair.Curve, tksGx, tksGy))
	b.Write(elliptic.Marshal(keypair.Curve, tksHx, tksHy))
	h2 := keypair.HashToField(b.Bytes())

	var buf bytes.Buffer
	buf.Write(make([]byte, 32-len(h2.Bytes())))
	buf.Write(h2.Bytes())

	if !hmac.Equal(s, buf.Bytes()) {
		return nilIndex, ErrInvalidVRF
	}
	return sha256.Sum256(vrf), nil
}
