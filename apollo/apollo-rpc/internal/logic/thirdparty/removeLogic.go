package thirdpartylogic

import (
	"context"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRemoveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveLogic {
	return &RemoveLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Remove 移除第三方账号
func (l *RemoveLogic) Remove(in *apollo.ThirdPartyRemoveReq) (*apollo.Empty, error) {
	err := l.svcCtx.ThirdPartyModel.Delete(l.ctx, in.ThirdPartyId, in.UserId)
	if err != nil {
		return nil, err
	}

	return &apollo.Empty{}, nil
}
