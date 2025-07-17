package passkeys

import (
	"context"
	"fmt"
	"jian-unified-system/apollo/apollo-rpc/apollo"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegFinishLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegFinishLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegFinishLogic {
	return &RegFinishLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegFinishLogic) RegFinish(req *types.RegFinishReq) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line
	// 1. 参数校验
	if len(req.SessionID) == 0 || len(req.Credential) == 0 {
		return nil, fmt.Errorf("参数错误: session_id和credential不能为空")
	}

	// 2. 从Redis获取SessionData
	sessionData, err := l.svcCtx.Redis.GetCtx(l.ctx, req.SessionID)
	if err != nil {
		l.Logger.Errorf("获取会话失败: key=%s, err=%v", req.SessionID, err)
		return nil, fmt.Errorf("会话已过期或不存在")
	}

	// 3. 调用gRPC服务完成注册
	_, err = l.svcCtx.ApolloRpc.FinishRegistration(l.ctx, &apollo.FinishRegistrationReq{
		SessionData:    []byte(sessionData),
		CredentialJson: []byte(req.Credential),
	})
	if err != nil {
		l.Logger.Errorf("gRPC调用失败: err=%v", err)
		return nil, fmt.Errorf("注册验证失败: %v", err)
	}

	// 4. 清理会话数据
	if _, err := l.svcCtx.Redis.DelCtx(l.ctx, req.SessionID); err != nil {
		l.Logger.Errorf("删除会话失败: key=%s, err=%v", req.SessionID, err)
		// 此处不返回错误，因为主流程已成功
	}

	// 5. 返回成功响应
	return &types.BaseResponse{
		Code:    200,
		Message: "注册成功",
	}, nil
}
