package logic

import (
	"context"
	"encoding/binary"
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
	// todo: add your logic here and delete this line
	// 1. 反序列化SessionData（使用完整包路径）
	var session webauthn.SessionData
	if err := json.Unmarshal(in.SessionData, &session); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid session data: "+err.Error())
	}

	// 2. 创建模拟HTTP请求
	httpRequest, err := createCredentialRequest(in.CredentialJson)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid credential format")
	}

	// 2. 解析前端断言响应
	//assertionResponse, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(in.CredentialJson))
	//if err != nil {
	//	return nil, status.Error(codes.InvalidArgument, "invalid credential data: "+err.Error())
	//}

	// 3. 从session还原真实用户
	user := &types.User{
		ID: int64(binary.BigEndian.Uint64(session.UserID)), // 从session还原
		// 注意：此处不需要其他字段，因FinishLogin只验证签名
	}

	// 4. 验证断言
	credential, err := l.svcCtx.WebAuthn.FinishLogin(user, session, httpRequest)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "assertion verification failed: "+err.Error())
	}

	// 5. 返回验证结果
	return &passkeys.FinishLoginResp{
		CredentialId: credential.ID,
		UserId:       user.ID, // 从SessionData还原
	}, nil
}
