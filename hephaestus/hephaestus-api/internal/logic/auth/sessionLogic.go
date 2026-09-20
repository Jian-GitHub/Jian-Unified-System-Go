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

type SessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SessionLogic {
	return &SessionLogic{ctx, svcCtx}
}
func (l *SessionLogic) Session(req *types.Empty) (*types.SessionResp, error) {
	input, e := access.Request[income.SessionRequest](l.ctx, req)
	if e != nil {
		return nil, e
	}
	v, e := l.svcCtx.Auth.Session(l.ctx, input)
	if e != nil {
		return nil, e
	}
	return transport.Convert[types.SessionResp](v)
}
