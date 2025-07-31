package types

import (
	"encoding/binary"
	"github.com/go-webauthn/webauthn/webauthn"
)

// WebauthnUser 统一用于注册/登录流程的User实现
// 实现webauthn.User接口（所有字段由API传入）
type WebauthnUser struct {
	ID          int64
	Name        string
	DisplayName string
	Credentials []webauthn.Credential // API传入的完整凭证
}

func (u *WebauthnUser) WebAuthnID() []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(u.ID))
	return buf
}
func (u *WebauthnUser) WebAuthnName() string        { return u.Name }
func (u *WebauthnUser) WebAuthnDisplayName() string { return u.DisplayName }
func (u *WebauthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials // 关键变更：使用API传入的凭证
}
