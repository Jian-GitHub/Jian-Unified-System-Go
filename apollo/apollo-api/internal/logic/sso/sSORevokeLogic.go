// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package sso

import (
	"context"
	"jian-unified-system/apollo/apollo-api/internal/logic/response"
	"jian-unified-system/apollo/apollo-rpc/apollo"

	"google.golang.org/grpc/metadata"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SSORevokeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSSORevokeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SSORevokeLogic {
	return &SSORevokeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SSORevokeLogic) SSORevoke(req *types.SSOTokenReq) (resp *types.OkResp, err error) {
	ctx := metadata.AppendToOutgoingContext(l.ctx, "x-apollo-sso-service", l.svcCtx.Config.SSO.RPCSecret)
	_, e := l.svcCtx.SSO.Revoke(ctx, &apollo.SSOTokenReq{ClientId: req.ClientId, ClientSecret: req.ClientSecret, Token: req.Token})
	if e != nil {
		return nil, e
	}
	return &types.OkResp{BaseResponse: response.OK()}, nil
}
