package accountlogic

import (
	"context"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
	"jian-unified-system/apollo/apollo-rpc/internal/logic/response"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateNameLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateNameLogic {
	return &UpdateNameLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateNameLogic) UpdateName(in *apollo.UpdateNameReq) (*apollo.Empty, error) {
	if in == nil {
		return nil, response.Error(identity.ErrInvalid)
	}
	if err := l.svcCtx.Account.UpdateName(l.ctx, in.UserId, in.GivenName, in.MiddleName, in.FamilyName); err != nil {
		return nil, response.Error(err)
	}
	return &apollo.Empty{}, nil
}
