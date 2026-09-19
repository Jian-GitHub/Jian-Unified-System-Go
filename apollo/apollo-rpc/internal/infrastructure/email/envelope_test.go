package email

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/cloudflare/circl/kem/mlkem/mlkem768"
)

func TestLegacyEnvelope(t *testing.T) {
	scheme := mlkem768.Scheme()
	pub, priv, err := scheme.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := pub.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	e, err := New(base64.StdEncoding.EncodeToString(raw))
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := e.Encrypt("user@example.com")
	if err != nil {
		t.Fatal(err)
	}
	payload, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatal(err)
	}
	// Independently decode the legacy wire layout, without importing the old utility package.
	const nonceSize = 12
	end := nonceSize + scheme.CiphertextSize()
	shared, err := scheme.Decapsulate(priv, payload[nonceSize:end])
	if err != nil {
		t.Fatal(err)
	}
	key := sha256.Sum256(shared)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		t.Fatal(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := gcm.Open(nil, payload[:nonceSize], payload[end:], nil)
	if err != nil || string(plain) != "user@example.com" {
		t.Fatalf("envelope mismatch: %v", err)
	}
	payload[len(payload)-1] ^= 1
	if _, err = gcm.Open(nil, payload[:nonceSize], payload[end:], nil); err == nil {
		t.Fatal("accepted tampered ciphertext")
	}
	if _, err = New("invalid"); err == nil {
		t.Fatal("accepted invalid key")
	}
	if LookupKey("User@example.com") == LookupKey("user@example.com") {
		t.Fatal("lookup silently folds case")
	}
}

func TestDecryptRejectsMalformedAndTamperedEnvelopes(t *testing.T) {
	pub, priv, err := mlkem768.Scheme().GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	p, _ := pub.MarshalBinary()
	s, _ := priv.MarshalBinary()
	codec, err := NewWithPrivate(base64.StdEncoding.EncodeToString(p), base64.StdEncoding.EncodeToString(s))
	if err != nil {
		t.Fatal(err)
	}
	value, err := codec.Encrypt("user@example.com")
	if err != nil {
		t.Fatal(err)
	}
	plain, err := codec.Decrypt(value)
	if err != nil || plain != "user@example.com" {
		t.Fatal("round trip failed")
	}
	raw, _ := base64.StdEncoding.DecodeString(value)
	raw[len(raw)-1] ^= 1
	for _, input := range []string{"", "invalid", base64.StdEncoding.EncodeToString([]byte{1}), base64.StdEncoding.EncodeToString(raw)} {
		if _, err = codec.Decrypt(input); err == nil {
			t.Fatal("invalid envelope accepted")
		}
	}
	_, other, err := mlkem768.Scheme().GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	key, _ := other.MarshalBinary()
	if _, err = NewWithPrivate(base64.StdEncoding.EncodeToString(p), base64.StdEncoding.EncodeToString(key)); err == nil {
		t.Fatal("mismatched key pair accepted")
	}
}
