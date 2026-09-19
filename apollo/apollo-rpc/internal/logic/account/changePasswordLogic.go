package accountlogic

import (
	"context"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
	"jian-unified-system/apollo/apollo-rpc/internal/logic/response"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChangePasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChangePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangePasswordLogic {
	return &ChangePasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ChangePasswordLogic) ChangePassword(in *apollo.ChangePasswordReq) (*apollo.Empty, error) {
	if in == nil {
		return nil, response.Error(identity.ErrInvalid)
	}
	if err := l.svcCtx.Account.ChangePassword(l.ctx, in.UserId, in.CurrentPassword, in.NewPassword, in.SignOutEverywhere); err != nil {
		return nil, response.Error(err)
	}
	return &apollo.Empty{}, nil
}
