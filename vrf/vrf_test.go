package vrf

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"github.com/Evanesco-Labs/miner/keypair"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestECMarshall(t *testing.T) {
	key, err := ecdsa.GenerateKey(keypair.Curve, rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	privateKey, err := keypair.NewPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}

	privBytes := elliptic.Marshal(keypair.Curve, privateKey.X, privateKey.Y)

	x, y := elliptic.Unmarshal(keypair.Curve, privBytes)
	if x == nil || y == nil {
		t.Fatal("nil")
	}

	assert.Equal(t, privateKey.X.Cmp(x), 0)
	assert.Equal(t, privateKey.Y.Cmp(y), 0)
}

func TestVrf(t *testing.T) {
	_, key := keypair.GenerateKeyPair()
	privateKey, err := keypair.NewPrivateKey(key.PrivateKey)
	if err != nil {
		t.Fatal(err)
	}

	publicKey, err := keypair.NewPublicKey(&key.PublicKey)

	message := []byte("go evanesco")
	rExp, proof := Evaluate(privateKey, message)

	if proof == nil {
		t.Fatalf("Evaluate fatal")
	}

	r, err := ProofToHash(publicKey, message, proof)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, rExp, r, "Vrf hash should be the same.")
}
