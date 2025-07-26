package passkeys

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-zero/core/jsonx"
	"jian-unified-system/apollo/apollo-rpc/passkeys"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginFinishLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginFinishLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginFinishLogic {
	return &LoginFinishLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginFinishLogic) LoginFinish(req *types.LoginFinishReq) (resp *types.LoginFinishResp, err error) {
	// 1. 参数校验
	if req.SessionID == "" || req.Assertion == "" {
		return nil, fmt.Errorf("参数不完整")
	}

	// 2. 解析前端传来的JSON断言
	var assertion struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		RawID    string `json:"rawId"`
		Response struct {
			ClientDataJSON    string `json:"clientDataJSON"`
			AuthenticatorData string `json:"authenticatorData"`
			Signature         string `json:"signature"`
			UserHandle        string `json:"userHandle"`
		} `json:"response"`
		ClientExtensionResults map[string]interface{} `json:"clientExtensionResults"`
	}

	if err := json.Unmarshal([]byte(req.Assertion), &assertion); err != nil {
		return nil, fmt.Errorf("解析断言失败: %w", err)
	}

	// 2. 获取SessionData
	sessionData, err := l.svcCtx.Redis.GetCtx(l.ctx, req.SessionID)
	if err != nil {
		l.Logger.Errorf("获取会话失败: key=%s, err=%v", req.SessionID, err)
		return nil, fmt.Errorf("会话已过期，请重新登录")
	}

	// 3. 调用gRPC验证
	_, err = l.svcCtx.ApolloRpc.FinishLogin(l.ctx, &passkeys.FinishLoginReq{
		SessionData:    []byte(sessionData),
		CredentialJson: []byte(req.Assertion),
		//CredentialJson: param,
	})
	if err != nil {
		l.Logger.Errorf("登录验证失败: err=%v", err)
		return nil, fmt.Errorf("身份验证失败")
	}

	// TODO: 4. 生成业务Token
	token := "token here"
	marshalToString, err := jsonx.MarshalToString(token)
	if err != nil {
		return nil, err
	}

	// 5. 清理会话
	_, _ = l.svcCtx.Redis.DelCtx(l.ctx, req.SessionID)

	return &types.LoginFinishResp{
		BaseResponse: types.BaseResponse{Code: 200, Message: "success"},
		Token:        marshalToString,
	}, nil
}
