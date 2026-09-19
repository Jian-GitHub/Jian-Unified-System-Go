package accountlogic

import (
	"context"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
	"jian-unified-system/apollo/apollo-rpc/internal/logic/response"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserInfoLogic {
	return &UserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UserInfoLogic) UserInfo(in *apollo.UserInfoReq) (*apollo.UserInfoResp, error) {
	if in == nil {
		return nil, response.Error(identity.ErrInvalid)
	}
	u, err := l.svcCtx.Account.User(l.ctx, in.UserId)
	if err != nil {
		return nil, response.Error(err)
	}
	body, err := response.LegacyUser(u)
	if err != nil {
		return nil, response.Error(err)
	}
	return &apollo.UserInfoResp{UserBytes: body, Profile: response.Profile(u)}, nil
}
