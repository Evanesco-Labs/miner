package keypair

import (
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	S256   = crypto.S256()
	Curve  = S256
	Params = Curve.Params()
)

var (
	ErrUnmarshallPrivateKey = errors.New("unmarshall private key from bytes error")
	ErrPointNotOnCurve      = errors.New("point is not on the curve")
	ErrPointNotCurveParam   = errors.New("point param not same as curve")
)

type Key struct {
	Address    common.Address
	PrivateKey PrivateKey
}

type PublicKey struct {
	*ecdsa.PublicKey
}

type PrivateKey struct {
	*ecdsa.PrivateKey
}

func (k *PrivateKey) Public() *PublicKey {
	return &PublicKey{&k.PublicKey}
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

func NewKey(sk *ecdsa.PrivateKey) (Key, error) {
	privKey, err := NewPrivateKey(sk)
	if err != nil {
		return Key{}, err
	}
	return Key{
		Address:    crypto.PubkeyToAddress(sk.PublicKey),
		PrivateKey: *privKey,
	}, nil
}
