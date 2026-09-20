// Code scaffolded by goctl 1.9.2. Safe to edit: thin RPC adapter.
package records

import (
	"context"
	"jian-unified-system/hephaestus/hephaestus-api/internal/access"
	"jian-unified-system/hephaestus/hephaestus-api/internal/svc"
	"jian-unified-system/hephaestus/hephaestus-api/internal/types"
	"jian-unified-system/hephaestus/hephaestus-rpc/income"
	"jian-unified-system/hephaestus/internal/transport"
)

type ClearOverrideLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewClearOverrideLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearOverrideLogic {
	return &ClearOverrideLogic{ctx, svcCtx}
}
func (l *ClearOverrideLogic) ClearOverride(req *types.ClearOverrideReq) (*types.Record, error) {
	input, e := access.Request[income.ClearOverrideRequest](l.ctx, req)
	if e != nil {
		return nil, e
	}
	v, e := l.svcCtx.Records.ClearOverride(l.ctx, input)
	if e != nil {
		return nil, e
	}
	return transport.Convert[types.Record](v)
}
