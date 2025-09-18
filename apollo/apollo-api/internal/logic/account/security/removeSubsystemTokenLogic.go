package security

import (
	"context"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveSubsystemTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRemoveSubsystemTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveSubsystemTokenLogic {
	return &RemoveSubsystemTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RemoveSubsystemTokenLogic) RemoveSubsystemToken(req *types.RemoveSubsystemTokenReq) (resp *types.RemoveSubsystemTokenResp, err error) {
	// todo: add your logic here and delete this line

	return
}
