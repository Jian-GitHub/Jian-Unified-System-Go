package passkey

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"net/http"

	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

type Verifier struct{ web *webauthn.WebAuthn }

func New(id, name string, origins []string) (*Verifier, error) {
	w, err := webauthn.New(&webauthn.Config{RPID: id, RPDisplayName: name, RPOrigins: origins, AuthenticatorSelection: protocol.AuthenticatorSelection{ResidentKey: protocol.ResidentKeyRequirementRequired, UserVerification: protocol.VerificationRequired}})
	if err != nil {
		return nil, err
	}
	return &Verifier{web: w}, nil
}

type user struct {
	source      identity.CeremonyUser
	credentials []webauthn.Credential
}

func (u user) WebAuthnID() []byte {
	v := make([]byte, 8)
	binary.BigEndian.PutUint64(v, uint64(u.source.ID))
	return v
}
func (u user) WebAuthnName() string                       { return u.source.Name }
func (u user) WebAuthnDisplayName() string                { return u.source.Name }
func (u user) WebAuthnCredentials() []webauthn.Credential { return u.credentials }
func decodeUser(in identity.CeremonyUser) (user, error) {
	u := user{source: in}
	for _, data := range in.Credentials {
		var c webauthn.Credential
		if err := json.Unmarshal(data, &c); err != nil {
			return u, err
		}
		u.credentials = append(u.credentials, c)
	}
	return u, nil
}
func (v *Verifier) BeginRegistration(in identity.CeremonyUser) ([]byte, []byte, error) {
	u, err := decodeUser(in)
	if err != nil {
		return nil, nil, err
	}
	excluded := []protocol.CredentialDescriptor{}
	for _, c := range u.credentials {
		excluded = append(excluded, c.Descriptor())
	}
	options, session, err := v.web.BeginRegistration(u, webauthn.WithExclusions(excluded))
	if err != nil {
		return nil, nil, err
	}
	a, err := json.Marshal(options)
	if err != nil {
		return nil, nil, err
	}
	b, err := json.Marshal(session)
	return a, b, err
}
func credential(c *webauthn.Credential) (identity.VerifiedCredential, error) {
	if c.Authenticator.CloneWarning {
		return identity.VerifiedCredential{}, errors.New("credential counter regression")
	}
	data, err := json.Marshal(c)
	return identity.VerifiedCredential{ID: base64.RawURLEncoding.EncodeToString(c.ID), Data: data, Count: c.Authenticator.SignCount}, err
}
func (v *Verifier) FinishRegistration(in identity.CeremonyUser, session, payload []byte) (identity.VerifiedCredential, error) {
	u, err := decodeUser(in)
	if err != nil {
		return identity.VerifiedCredential{}, err
	}
	var s webauthn.SessionData
	if err = json.Unmarshal(session, &s); err != nil {
		return identity.VerifiedCredential{}, err
	}
	r, err := http.NewRequest(http.MethodPost, "http://localhost", bytes.NewReader(payload))
	if err != nil {
		return identity.VerifiedCredential{}, err
	}
	c, err := v.web.FinishRegistration(u, s, r)
	if err != nil {
		return identity.VerifiedCredential{}, err
	}
	return credential(c)
}
func (v *Verifier) BeginLogin() ([]byte, []byte, error) {
	o, s, err := v.web.BeginDiscoverableLogin(webauthn.WithUserVerification(protocol.VerificationRequired))
	if err != nil {
		return nil, nil, err
	}
	a, err := json.Marshal(o)
	if err != nil {
		return nil, nil, err
	}
	b, err := json.Marshal(s)
	return a, b, err
}
func (v *Verifier) FinishLogin(ctx context.Context, session, payload []byte, lookup func(string, []byte) (identity.CeremonyUser, error)) (identity.VerifiedCredential, error) {
	var s webauthn.SessionData
	if err := json.Unmarshal(session, &s); err != nil {
		return identity.VerifiedCredential{}, err
	}
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost", bytes.NewReader(payload))
	if err != nil {
		return identity.VerifiedCredential{}, err
	}
	c, err := v.web.FinishDiscoverableLogin(func(rawID, handle []byte) (webauthn.User, error) {
		in, err := lookup(base64.RawURLEncoding.EncodeToString(rawID), handle)
		if err != nil {
			return nil, err
		}
		return decodeUser(in)
	}, s, r)
	if err != nil {
		return identity.VerifiedCredential{}, err
	}
	return credential(c)
}
