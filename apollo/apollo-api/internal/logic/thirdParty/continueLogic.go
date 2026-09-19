// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package thirdParty

import (
	"context"
	"jian-unified-system/apollo/apollo-api/internal/infrastructure/geo"
	"jian-unified-system/apollo/apollo-api/internal/logic/response"
	"jian-unified-system/apollo/apollo-rpc/apollo"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ContinueLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewContinueLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ContinueLogic {
	return &ContinueLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ContinueLogic) Continue(req *types.ProviderReq, locale ...string) (resp *types.AuthorizationResp, err error) {
	resolved := geo.UnknownCountry
	if len(locale) > 0 && locale[0] != "" {
		resolved = locale[0]
	}
	r, err := l.svcCtx.ThirdParty.StartAuthorization(l.ctx, &apollo.StartAuthorizationReq{Provider: req.Provider, UserId: l.svcCtx.Sessions.NextID(), Locale: resolved, Language: l.svcCtx.Config.DefaultLanguage})
	if err != nil {
		return nil, err
	}
	return &types.AuthorizationResp{BaseResponse: response.OK(), Data: types.AuthorizationData{Url: r.Url}}, nil
}
