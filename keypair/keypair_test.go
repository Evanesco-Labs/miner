package keypair

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMarshall(t *testing.T) {
	pub, priv := GenerateKeyPair()
	if pub == nil || priv == nil {
		t.Fatal("key generate error")
	}

	pubBytes := pub.Marshall()

	privBytes := priv.Marshall()

	var privR PrivateKey
	var pubR PublicKey

	err := privR.Unmarshall(privBytes)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, priv.X.Cmp(privR.X), 0)
	assert.Equal(t, priv.Y.Cmp(privR.Y), 0)

	err = pubR.Unmarshall(pubBytes)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, pub.X.Cmp(pubR.X), 0)
	assert.Equal(t, pub.Y.Cmp(pubR.Y), 0)
}
