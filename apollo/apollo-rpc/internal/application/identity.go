package application

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"

	"jian-unified-system/apollo/apollo-rpc/internal/domain/event"
)

// NoopPublisher is retained for source compatibility in adapter tests.
// Production wiring uses event.OutboxPublisher.
type NoopPublisher = event.NoopPublisher

// RandomID generates a positive, non-zero int64 suitable for application IDs.
func RandomID() (int64, error) {
	var raw [8]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return 0, err
	}
	n := int64(binary.BigEndian.Uint64(raw[:]) & 0x7fffffffffffffff)
	if n == 0 {
		return RandomID()
	}
	return n, nil
}

// OpaqueID generates an unguessable URL-safe identifier for short-lived
// authentication sessions.
func OpaqueID() (string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw[:]), nil
}
