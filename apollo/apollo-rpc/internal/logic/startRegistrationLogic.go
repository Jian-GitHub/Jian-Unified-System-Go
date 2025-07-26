package logic

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"jian-unified-system/apollo/apollo-rpc/internal/types"
	"jian-unified-system/apollo/apollo-rpc/passkeys"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type StartRegistrationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewStartRegistrationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StartRegistrationLogic {
	return &StartRegistrationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// StartRegistration 注册
func (l *StartRegistrationLogic) StartRegistration(in *apollo.StartRegistrationReq) (*apollo.StartRegistrationResp, error) {
	// 创建WebAuthn用户结构
	user := &types.User{
		ID:          in.UserId,
		Name:        in.UserName,
		DisplayName: in.DisplayName,
		Credentials: []webauthn.Credential{}, // 新用户无凭证
	}

	// 生成注册选项
	creation, session, err := l.svcCtx.WebAuthn.BeginRegistration(
		user,
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
	)
	if err != nil {
		l.Logger.Error("WebAuthn.BeginRegistration failed: ", err)
		return nil, status.Error(codes.Internal, "failed to generate challenge")
	}

	// 4. 返回CredentialCreation的JSON
	creationJson, err := json.Marshal(creation)
	if err != nil {
		return nil, errors.New("failed to marshal options")
	}

	// 序列化SessionData（API需存储此数据）
	sessionData, err := json.Marshal(session)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to marshal session")
	}

	return &passkeys.StartRegistrationResp{
		OptionsJson: creationJson,
		SessionData: sessionData,
	}, nil
}
