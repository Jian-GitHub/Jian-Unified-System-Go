// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package sso

import (
	"context"
	"jian-unified-system/apollo/apollo-api/internal/application"
	"jian-unified-system/apollo/apollo-api/internal/logic/response"
	"jian-unified-system/apollo/apollo-rpc/apollo"

	"google.golang.org/grpc/metadata"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SSOAuthorizeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSSOAuthorizeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SSOAuthorizeLogic {
	return &SSOAuthorizeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SSOAuthorizeLogic) SSOAuthorize(req *types.SSOAuthorizeReq) (resp *types.SSOCodeResp, err error) {
	ctx := metadata.AppendToOutgoingContext(l.ctx, "x-apollo-sso-service", l.svcCtx.Config.SSO.RPCSecret)
	id, e := response.Subject(l.ctx)
	if e != nil {
		return nil, e
	}
	version, ok := l.ctx.Value("authVersion").(int64)
	if !ok {
		return nil, application.ErrCredentials
	}
	v, e := l.svcCtx.SSO.Authorize(ctx, &apollo.SSOAuthorizeReq{UserId: id, AuthVersion: version, ClientId: req.ClientId, RedirectUri: req.RedirectUri, CodeChallenge: req.CodeChallenge})
	if e != nil {
		return nil, e
	}
	return &types.SSOCodeResp{Code: v.Code}, nil
}
