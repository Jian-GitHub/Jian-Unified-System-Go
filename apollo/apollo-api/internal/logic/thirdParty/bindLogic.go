// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package thirdParty

import (
	"context"
	"jian-unified-system/apollo/apollo-api/internal/logic/response"
	"jian-unified-system/apollo/apollo-rpc/apollo"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BindLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBindLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindLogic {
	return &BindLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BindLogic) Bind(req *types.ProviderReq) (resp *types.AuthorizationResp, err error) {
	id, err := response.Subject(l.ctx)
	if err != nil {
		return nil, err
	}
	r, err := l.svcCtx.ThirdParty.StartAuthorization(l.ctx, &apollo.StartAuthorizationReq{Provider: req.Provider, UserId: id, Bind: true, Locale: l.svcCtx.Config.DefaultLocale, Language: l.svcCtx.Config.DefaultLanguage})
	if err != nil {
		return nil, err
	}
	return &types.AuthorizationResp{BaseResponse: response.OK(), Data: types.AuthorizationData{Url: r.Url}}, nil
}
