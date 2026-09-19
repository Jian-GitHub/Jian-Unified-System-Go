package integration

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"testing"

	"github.com/fxamacker/cbor/v2"
)

// A real ES256 virtual authenticator: server verification is never stubbed.
type authenticator struct {
	key        *ecdsa.PrivateKey
	id, handle []byte
}

func encoded(v []byte) string { return base64.RawURLEncoding.EncodeToString(v) }
func newAuthenticator(t *testing.T) *authenticator {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	id := make([]byte, 32)
	if _, err = rand.Read(id); err != nil {
		t.Fatal(err)
	}
	return &authenticator{key: key, id: id}
}
func options(t *testing.T, raw string) map[string]any {
	t.Helper()
	var value map[string]any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		t.Fatal(err)
	}
	return value["publicKey"].(map[string]any)
}
func authData(flags byte, count uint32) []byte {
	hash := sha256.Sum256([]byte("localhost"))
	data := append([]byte{}, hash[:]...)
	data = append(data, flags, 0, 0, 0, 0)
	binary.BigEndian.PutUint32(data[33:], count)
	return data
}
func (a *authenticator) registration(t *testing.T, raw string) string {
	t.Helper()
	opts := options(t, raw)
	var err error
	a.handle, err = base64.RawURLEncoding.DecodeString(opts["user"].(map[string]any)["id"].(string))
	if err != nil {
		t.Fatal(err)
	}
	client, _ := json.Marshal(map[string]any{"type": "webauthn.create", "challenge": opts["challenge"], "origin": "http://localhost:3000"})
	cose, err := cbor.Marshal(map[int]any{1: 2, 3: -7, -1: 1, -2: a.key.X.FillBytes(make([]byte, 32)), -3: a.key.Y.FillBytes(make([]byte, 32))})
	if err != nil {
		t.Fatal(err)
	}
	data := authData(0x45, 0)
	data = append(data, make([]byte, 16)...)
	data = append(data, byte(len(a.id)>>8), byte(len(a.id)))
	data = append(data, a.id...)
	data = append(data, cose...)
	attestation, err := cbor.Marshal(map[string]any{"fmt": "none", "attStmt": map[string]any{}, "authData": data})
	if err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(map[string]any{"id": encoded(a.id), "rawId": encoded(a.id), "type": "public-key", "response": map[string]any{"clientDataJSON": encoded(client), "attestationObject": encoded(attestation), "transports": []string{"internal"}}, "clientExtensionResults": map[string]any{}})
	return string(payload)
}
func (a *authenticator) assertion(t *testing.T, raw string, count uint32, origin string, corrupt bool) string {
	t.Helper()
	opts := options(t, raw)
	client, _ := json.Marshal(map[string]any{"type": "webauthn.get", "challenge": opts["challenge"], "origin": origin})
	data := authData(0x05, count)
	clientHash := sha256.Sum256(client)
	signed := append(append([]byte{}, data...), clientHash[:]...)
	hash := sha256.Sum256(signed)
	sig, err := ecdsa.SignASN1(rand.Reader, a.key, hash[:])
	if err != nil {
		t.Fatal(err)
	}
	if corrupt {
		sig[len(sig)-1] ^= 1
	}
	payload, _ := json.Marshal(map[string]any{"id": encoded(a.id), "rawId": encoded(a.id), "type": "public-key", "response": map[string]any{"clientDataJSON": encoded(client), "authenticatorData": encoded(data), "signature": encoded(sig), "userHandle": encoded(a.handle)}, "clientExtensionResults": map[string]any{}})
	return string(payload)
}
