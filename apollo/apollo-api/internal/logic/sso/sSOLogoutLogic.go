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

type SSOLogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSSOLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SSOLogoutLogic {
	return &SSOLogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SSOLogoutLogic) SSOLogout(req *types.Empty) (resp *types.OkResp, err error) {
	id, e := response.Subject(l.ctx)
	if e != nil {
		return nil, e
	}
	ctx := metadata.AppendToOutgoingContext(l.ctx, "x-apollo-sso-service", l.svcCtx.Config.SSO.RPCSecret)
	_, e = l.svcCtx.SSO.Logout(ctx, &apollo.SSOLogoutReq{UserId: id})
	if e != nil {
		return nil, e
	}
	return &types.OkResp{BaseResponse: response.OK()}, nil
}
