package logic

import (
	"bytes"
	"context"
	"database/sql"
	"jian-unified-system/apollo/apollo-rpc/internal/model"
	"jian-unified-system/apollo/apollo-rpc/internal/types"
	"jian-unified-system/apollo/apollo-rpc/passkeys"
	"net/http"
	"net/http/httptest"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"encoding/json"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FinishRegistrationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFinishRegistrationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FinishRegistrationLogic {
	return &FinishRegistrationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// FinishRegistration 注册第二步 - 完成
func (l *FinishRegistrationLogic) FinishRegistration(in *apollo.FinishRegistrationReq) (*apollo.FinishRegistrationResp, error) {
	// todo: add your logic here and delete this line
	// 1. 参数校验
	if len(in.CredentialJson) == 0 || len(in.SessionData) == 0 {
		return nil, status.Error(codes.InvalidArgument, "missing required fields")
	}

	// 2. 反序列化SessionData
	var session webauthn.SessionData
	if err := json.Unmarshal(in.SessionData, &session); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid session data")
	}

	// 3. 创建模拟HTTP请求
	req, err := createCredentialRequest(in.CredentialJson)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid credential format")
	}

	user := &types.User{ID: in.UserId}
	// 4. 验证凭证
	credential, err := l.svcCtx.WebAuthn.FinishRegistration(
		user,
		session,
		req,
	)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "webauthn verification failed: "+err.Error())
	}

	// 1. 保存Passkeys记录（完全匹配你的结构体）
	passkey := model.Passkeys{
		Handle: sql.NullString{String: string(user.WebAuthnID()), Valid: true}, // 注意：这里按你的要求转为string
		UserId: 12,                                                             //l.svcCtx.snow,
	}

	println(passkey.Id)

	return &passkeys.FinishRegistrationResp{
		CredentialId: credential.ID,
		PublicKey:    credential.PublicKey,
	}, nil
}

// 辅助函数：将JSON转换为http.Request
func createCredentialRequest(jsonData []byte) (*http.Request, error) {
	req := httptest.NewRequest("POST", "/", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}
