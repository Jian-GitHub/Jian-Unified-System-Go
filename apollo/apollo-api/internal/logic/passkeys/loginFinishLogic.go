package passkeys

import (
	"context"
	"fmt"
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
	// todo: add your logic here and delete this line
	// 1. 参数校验
	if req.SessionID == "" || req.Assertion == "" {
		return nil, fmt.Errorf("参数不完整")
	}

	// 2. 获取SessionData
	sessionData, err := l.svcCtx.Redis.GetCtx(l.ctx, req.SessionID)
	if err != nil {
		l.Logger.Errorf("获取会话失败: key=%s, err=%v", req.SessionID, err)
		return nil, fmt.Errorf("会话已过期，请重新登录")
	}

	// 3. 调用gRPC验证
	_, err = l.svcCtx.PasskeysRpc.FinishLogin(l.ctx, &passkeys.FinishLoginReq{
		SessionData:    []byte(sessionData),
		CredentialJson: []byte(req.Assertion),
	})
	if err != nil {
		l.Logger.Errorf("登录验证失败: err=%v", err)
		return nil, fmt.Errorf("身份验证失败")
	}

	// 4. 生成业务Token
	//token, err := l.svcCtx.Auth.GenerateToken(finishResp.UserId)
	//if err != nil {
	//	l.Logger.Errorf("Token生成失败: user_id=%x, err=%v", finishResp.UserId, err)
	//	return nil, fmt.Errorf("系统错误")
	//}
	token := "token here"

	// 5. 清理会话
	_, _ = l.svcCtx.Redis.DelCtx(l.ctx, req.SessionID)

	return &types.LoginFinishResp{
		BaseResponse: types.BaseResponse{Code: 200, Message: "success"},
		Token:        token,
	}, nil
}
