package logic

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"jian-unified-system/apollo/apollo-rpc/internal/types"
	"jian-unified-system/apollo/apollo-rpc/passkeys"
	"strings"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type FinishLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFinishLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FinishLoginLogic {
	return &FinishLoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// FinishLogin 登陆第二步 - 完成
func (l *FinishLoginLogic) FinishLogin(in *apollo.FinishLoginReq) (*apollo.FinishLoginResp, error) {
	// 1. 先Base64解码
	s := strings.Trim(string(in.SessionData), "\"")
	decoded, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "base64 decode session data failed: "+err.Error())
	}

	// 2. 反序列化SessionData（使用完整包路径）
	var session webauthn.SessionData
	if err := json.Unmarshal(decoded, &session); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid session data: "+err.Error())
	}

	// 3. 创建模拟HTTP请求
	httpRequest, err := createCredentialRequest(in.CredentialJson)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid credential format")
	}

	// 4. 验证断言
	credential, err := l.svcCtx.WebAuthn.FinishDiscoverableLogin(
		l.discoverableUserHandler,
		//user,
		session,
		httpRequest,
	)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "assertion verification failed: "+err.Error())
	}

	// 5. 返回验证结果
	return &passkeys.FinishLoginResp{
		CredentialId: credential.ID,
		UserId:       123, // 从SessionData还原
		//UserId:       int64(binary.BigEndian.Uint64(user.WebAuthnID())), // 从SessionData还原
	}, nil
}

// 实现 DiscoverableUserHandler
func (l *FinishLoginLogic) discoverableUserHandler(rawID, userHandle []byte) (webauthn.User, error) {
	// rawID - Passkeys ID
	// userHandle - User ID
	// 1. 将二进制 userHandle 转为整数 user_id
	if len(userHandle) < 8 {
		return nil, fmt.Errorf("invalid userHandle length")
	}

	// 3. 验证凭证属于该用户
	credentialID := base64.RawURLEncoding.EncodeToString(rawID)
	passkey, err := l.svcCtx.PasskeyModel.FindOne(
		l.ctx,
		credentialID,
	)
	if err != nil {
		return nil, fmt.Errorf("credential not associated with user")
	}

	// 解析 publicKey 和 credentialId
	credID, err := base64.RawURLEncoding.DecodeString(passkey.CredentialId)
	if err != nil {
		return nil, fmt.Errorf("invalid credentialId: %w", err)
	}

	pubKey, err := base64.RawURLEncoding.DecodeString(passkey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("invalid publicKey: %w", err)
	}

	var trans []protocol.AuthenticatorTransport
	err = json.Unmarshal([]byte(passkey.Transports.String), &trans)
	if err != nil {
		return nil, err
	}

	// 构造 Credential
	credential := webauthn.Credential{
		ID:        credID,
		PublicKey: pubKey,
		Transport: trans,
		Authenticator: webauthn.Authenticator{
			SignCount: uint32(passkey.SignCount),
		},
		Flags: webauthn.CredentialFlags{
			BackupEligible: true,
			BackupState:    false,
		},
	}

	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(passkey.UserId))

	// 构造 LoginUser
	user := &types.User{
		ID:          passkey.UserId, // 用 user_id 作为 ID
		Name:        fmt.Sprintf("user_${Date.now()}@test.com"),
		DisplayName: passkey.DisplayName,
		Credentials: []webauthn.Credential{credential},
	}
	return user, nil
}
