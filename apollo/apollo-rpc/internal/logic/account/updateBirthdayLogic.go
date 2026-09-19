package accountlogic

import (
	"context"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
	"jian-unified-system/apollo/apollo-rpc/internal/logic/response"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateBirthdayLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateBirthdayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateBirthdayLogic {
	return &UpdateBirthdayLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateBirthdayLogic) UpdateBirthday(in *apollo.UpdateBirthdayReq) (*apollo.Empty, error) {
	if in == nil {
		return nil, response.Error(identity.ErrInvalid)
	}
	if err := l.svcCtx.Account.UpdateBirthday(l.ctx, in.UserId, in.Year, in.Month, in.Day); err != nil {
		return nil, response.Error(err)
	}
	return &apollo.Empty{}, nil
}
