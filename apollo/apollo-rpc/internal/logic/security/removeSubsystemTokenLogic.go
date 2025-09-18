package securitylogic

import (
	"context"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveSubsystemTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRemoveSubsystemTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveSubsystemTokenLogic {
	return &RemoveSubsystemTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 移除子系统令牌
func (l *RemoveSubsystemTokenLogic) RemoveSubsystemToken(in *apollo.RemoveSubsystemTokenReq) (*apollo.RemoveSubsystemTokenResp, error) {
	// todo: add your logic here and delete this line

	return &apollo.RemoveSubsystemTokenResp{}, nil
}
