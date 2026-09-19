package accountlogic

import (
	"context"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
	"jian-unified-system/apollo/apollo-rpc/internal/logic/response"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChangeNotificationEmailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChangeNotificationEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangeNotificationEmailLogic {
	return &ChangeNotificationEmailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ChangeNotificationEmailLogic) ChangeNotificationEmail(in *apollo.ChangeNotificationEmailReq) (*apollo.Empty, error) {
	if in == nil {
		return nil, response.Error(identity.ErrInvalid)
	}
	if err := l.svcCtx.Account.ChangeNotificationEmail(l.ctx, in.UserId, in.Email); err != nil {
		return nil, response.Error(err)
	}
	return &apollo.Empty{}, nil
}
