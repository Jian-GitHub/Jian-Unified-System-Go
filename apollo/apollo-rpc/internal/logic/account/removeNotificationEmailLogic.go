package accountlogic

import (
	"context"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
	"jian-unified-system/apollo/apollo-rpc/internal/logic/response"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveNotificationEmailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRemoveNotificationEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveNotificationEmailLogic {
	return &RemoveNotificationEmailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RemoveNotificationEmailLogic) RemoveNotificationEmail(in *apollo.RemoveNotificationEmailReq) (*apollo.Empty, error) {
	if in == nil {
		return nil, response.Error(identity.ErrInvalid)
	}
	if err := l.svcCtx.Account.RemoveNotificationEmail(l.ctx, in.UserId); err != nil {
		return nil, response.Error(err)
	}
	return &apollo.Empty{}, nil
}
