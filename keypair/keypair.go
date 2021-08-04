package keypair

import (
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"github.com/btcsuite/btcd/btcec"
)

var (
	S256   = btcec.S256()
	Curve  = S256
	Params = Curve.Params()
)

var (
	ErrUnmarshallPrivateKey = errors.New("unmarshall private key from bytes error")
	ErrPointNotOnCurve      = errors.New("point is not on the curve")
	ErrPointNotCurveParam   = errors.New("point param not same as curve")
)

type PublicKey struct {
	*ecdsa.PublicKey
}

type PrivateKey struct {
	*ecdsa.PrivateKey
}

func (k *PrivateKey) Public() *PublicKey {
	return &PublicKey{&k.PublicKey}
}

func (k *PublicKey) Marshall() []byte {
	return (*btcec.PublicKey)(k.PublicKey).SerializeCompressed()
}

func (k *PublicKey) Unmarshall(b []byte) error {
	pk, err := btcec.ParsePubKey(b, Curve)
	if err != nil {
		return err
	}
	k.PublicKey = (*ecdsa.PublicKey)(pk)
	return nil
}

func (k *PrivateKey) Marshall() []byte {
	return (*btcec.PrivateKey)(k.PrivateKey).Serialize()
}

func (k *PrivateKey) Unmarshall(b []byte) error {
	priv, pub := btcec.PrivKeyFromBytes(Curve, b)
	if priv == nil || pub == nil {
		return ErrUnmarshallPrivateKey
	}
	k.PrivateKey = (*ecdsa.PrivateKey)(priv)
	return nil
}

func GenerateKeyPair() (*PublicKey, *PrivateKey) {
	sk, err := ecdsa.GenerateKey(Curve, rand.Reader)
	if err != nil {
		return nil, nil
	}
	return &PublicKey{&sk.PublicKey}, &PrivateKey{sk}
}

func NewPublicKey(pk *ecdsa.PublicKey) (*PublicKey, error) {
	if *pk.Params() != *Params {
		return nil, ErrPointNotCurveParam
	}
	if !Curve.IsOnCurve(pk.X, pk.Y) {
		return nil, ErrPointNotOnCurve
	}
	return &PublicKey{pk}, nil
}

func NewPrivateKey(sk *ecdsa.PrivateKey) (*PrivateKey, error) {
	if *sk.Params() != *Params {
		return nil, ErrPointNotCurveParam
	}
	if !Curve.IsOnCurve(sk.X, sk.Y) {
		return nil, ErrPointNotOnCurve
	}
	return &PrivateKey{sk}, nil
}
