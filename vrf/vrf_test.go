package vrf

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"github.com/Evanesco-Labs/Miner/keypair"
	"github.com/stretchr/testify/assert"
	"testing"
)

var privTest = []byte{44, 213, 72, 196, 72, 214, 71, 131, 51, 90, 31, 207, 230, 165, 83, 129, 200, 77, 131, 220, 225, 7, 19, 120, 201, 229, 143, 206, 24, 145, 202, 159}

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
	var key keypair.PrivateKey
	err := key.Unmarshall(privTest)
	if err != nil {
		t.Fatal(err)
	}

	privateKey, err := keypair.NewPrivateKey(key.PrivateKey)
	if err != nil {
		t.Fatal(err)
	}

	publicKey, err := keypair.NewPublicKey(&key.PublicKey)

	message := []byte("go evanesco")
	hashExp, proof := Evaluate(privateKey, message)

	if proof == nil {
		t.Fatalf("Evaluate fatal")
	}

	hash, err := ProofToHash(publicKey, message, proof)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, hashExp, hash, "Vrf hash should be the same.")
}
