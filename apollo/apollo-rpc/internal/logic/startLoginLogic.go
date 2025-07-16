package logic

import (
	"context"
	"encoding/json"
	"github.com/go-webauthn/webauthn/webauthn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"jian-unified-system/apollo/apollo-rpc/internal/types"
	"jian-unified-system/apollo/apollo-rpc/passkeys"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type StartLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewStartLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StartLoginLogic {
	return &StartLoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 登录
func (l *StartLoginLogic) StartLogin(in *apollo.StartLoginReq) (*apollo.StartLoginResp, error) {
	// todo: add your logic here and delete this line
	// 1. 参数基础校验
	if !json.Valid(in.CredentialsJson) {
		return nil, status.Error(codes.InvalidArgument, "invalid credentials_json format")
	}

	// 2. 解析凭证列表
	var credentials []webauthn.Credential
	if err := json.Unmarshal(in.CredentialsJson, &credentials); err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to parse credentials: "+err.Error())
	}

	// 3. 构建WebAuthn用户结构
	user := &types.User{
		ID:          in.UserId,
		Name:        in.UserName,
		DisplayName: in.DisplayName,
		Credentials: credentials, // 使用API传入的凭证
	}

	// 4. 生成登录选项
	options, session, err := l.svcCtx.WebAuthn.BeginLogin(user)
	if err != nil {
		l.Logger.Errorf("WebAuthn BeginLogin failed: %v", err)
		return nil, status.Error(codes.Internal, "failed to generate login challenge")
	}

	// 5. 序列化响应数据
	optionsJson, err := json.Marshal(options.Response)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to marshal options")
	}

	sessionData, err := json.Marshal(session)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to marshal session")
	}

	return &passkeys.StartLoginResp{
		OptionsJson: optionsJson,
		SessionData: sessionData,
	}, nil
}

// 实现webauthn.User接口（精简版）
//type loginUser struct{}
//
//func (u *loginUser) WebAuthnID() []byte                         { return []byte("dummy-id") }
//func (u *loginUser) WebAuthnName() string                       { return "dummy" }
//func (u *loginUser) WebAuthnDisplayName() string                { return "Dummy User" }
//func (u *loginUser) WebAuthnCredentials() []webauthn.Credential { return []webauthn.Credential{} }
