package email

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"io"

	"github.com/cloudflare/circl/kem"
	"github.com/cloudflare/circl/kem/mlkem/mlkem768"
)

// LookupKey preserves the original Base64(SHA512(email + email)) format.
func LookupKey(address string) string {
	sum := sha512.Sum512([]byte(address + address))
	return base64.StdEncoding.EncodeToString(sum[:])
}

type Envelope struct {
	key     kem.PublicKey
	private kem.PrivateKey
}

func New(publicKey string) (*Envelope, error) {
	raw, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return nil, errors.New("invalid email public key")
	}
	key, err := mlkem768.Scheme().UnmarshalBinaryPublicKey(raw)
	if err != nil {
		return nil, errors.New("invalid email public key")
	}
	return &Envelope{key: key}, nil
}

func NewWithPrivate(publicKey, privateKey string) (*Envelope, error) {
	e, err := New(publicKey)
	if err != nil {
		return nil, err
	}
	raw, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return nil, errors.New("invalid email private key")
	}
	e.private, err = mlkem768.Scheme().UnmarshalBinaryPrivateKey(raw)
	if err != nil {
		return nil, errors.New("invalid email private key")
	}
	pub, err := e.private.Public().MarshalBinary()
	if err != nil {
		return nil, err
	}
	if base64.StdEncoding.EncodeToString(pub) != publicKey {
		return nil, errors.New("email key pair mismatch")
	}
	return e, nil
}

func (e *Envelope) Decrypt(encoded string) (string, error) {
	if e.private == nil || len(encoded) > 8192 {
		return "", errors.New("invalid email envelope")
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", errors.New("invalid email envelope")
	}
	const nonceSize = 12
	end := nonceSize + e.private.Scheme().CiphertextSize()
	if len(raw) < end+16 {
		return "", errors.New("invalid email envelope")
	}
	shared, err := e.private.Scheme().Decapsulate(e.private, raw[nonceSize:end])
	if err != nil {
		return "", errors.New("invalid email envelope")
	}
	key := sha256.Sum256(shared)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	plain, err := gcm.Open(nil, raw[:nonceSize], raw[end:], nil)
	if err != nil {
		return "", errors.New("invalid email envelope")
	}
	return string(plain), nil
}

// Encrypt preserves the legacy nonce | ML-KEM ciphertext | AES-GCM payload envelope.
// Only the public key is required for registration; no private key is loaded.
func (e *Envelope) Encrypt(address string) (string, error) {
	encapsulated, shared, err := e.key.Scheme().Encapsulate(e.key)
	if err != nil {
		return "", err
	}
	key := sha256.Sum256(shared)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	payload := append(nonce, encapsulated...)
	payload = append(payload, gcm.Seal(nil, nonce, []byte(address), nil)...)
	return base64.StdEncoding.EncodeToString(payload), nil
}
