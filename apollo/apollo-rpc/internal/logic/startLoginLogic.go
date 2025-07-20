package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"
	"jian-unified-system/apollo/apollo-rpc/passkeys"

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
func (l *StartLoginLogic) StartLogin() (*apollo.StartLoginResp, error) {
	// todo: add your logic here and delete this line
	// 不传用户名，不查用户，直接生成登录选项（无 allowCredentials）
	options, session, err := l.svcCtx.WebAuthn.BeginDiscoverableLogin(
	//webauthn.WithUserVerification(protocol.VerificationRequired),
	)
	if err != nil {
		l.Logger.Errorf("BeginLogin error: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	// 5. 序列化响应数据
	optionsJson, err := json.Marshal(options)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to marshal options")
	}

	sessionData, err := json.Marshal(session)
	fmt.Println("1 sessionData:", string(sessionData))
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
