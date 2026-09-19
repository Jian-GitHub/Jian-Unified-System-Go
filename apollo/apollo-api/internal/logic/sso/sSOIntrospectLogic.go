// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package sso

import (
	"context"
	"jian-unified-system/apollo/apollo-rpc/apollo"

	"google.golang.org/grpc/metadata"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SSOIntrospectLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSSOIntrospectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SSOIntrospectLogic {
	return &SSOIntrospectLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SSOIntrospectLogic) SSOIntrospect(req *types.SSOTokenReq) (resp *types.SSOSessionResp, err error) {
	ctx := metadata.AppendToOutgoingContext(l.ctx, "x-apollo-sso-service", l.svcCtx.Config.SSO.RPCSecret)
	v, e := l.svcCtx.SSO.Introspect(ctx, &apollo.SSOTokenReq{ClientId: req.ClientId, ClientSecret: req.ClientSecret, Token: req.Token})
	return sessionResponse(v, e)
}
