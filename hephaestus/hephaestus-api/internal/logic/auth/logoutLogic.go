// Code scaffolded by goctl 1.9.2. Safe to edit: thin RPC adapter.
package auth

import (
	"context"
	"jian-unified-system/hephaestus/hephaestus-api/internal/access"
	"jian-unified-system/hephaestus/hephaestus-api/internal/svc"
	"jian-unified-system/hephaestus/hephaestus-api/internal/types"
	"jian-unified-system/hephaestus/hephaestus-rpc/income"
	"jian-unified-system/hephaestus/internal/transport"
)

type LogoutLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{ctx, svcCtx}
}
func (l *LogoutLogic) Logout(req *types.Empty) (*types.Empty, error) {
	input, e := access.Request[income.LogoutRequest](l.ctx, req)
	if e != nil {
		return nil, e
	}
	v, e := l.svcCtx.Auth.Logout(l.ctx, input)
	if e != nil {
		return nil, e
	}
	l.svcCtx.Access.ClearSession(l.ctx)
	return transport.Convert[types.Empty](v)
}
