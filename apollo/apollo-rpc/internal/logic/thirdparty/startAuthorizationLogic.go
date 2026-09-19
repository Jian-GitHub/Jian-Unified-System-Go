package thirdpartylogic

import (
	"context"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/account"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
	"jian-unified-system/apollo/apollo-rpc/internal/logic/response"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type StartAuthorizationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewStartAuthorizationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StartAuthorizationLogic {
	return &StartAuthorizationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *StartAuthorizationLogic) StartAuthorization(in *apollo.StartAuthorizationReq) (*apollo.StartAuthorizationResp, error) {
	if in == nil {
		return nil, response.Error(identity.ErrInvalid)
	}
	in.Locale = account.NormalizeLocale(in.Locale)
	profile, err := account.NewMinimalProfile(in.UserId, in.Locale, in.Language)
	if err != nil {
		return nil, response.Error(err)
	}
	url, err := l.svcCtx.OAuth.StartAuthorization(l.ctx, in.Provider, profile, in.Bind)
	if err != nil {
		return nil, response.Error(err)
	}
	return &apollo.StartAuthorizationResp{Url: url}, nil
}
