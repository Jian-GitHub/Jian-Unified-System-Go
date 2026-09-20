// Code scaffolded by goctl 1.9.2. Safe to edit: thin RPC adapter.
package health

import (
	"context"
	"jian-unified-system/hephaestus/hephaestus-api/internal/access"
	"jian-unified-system/hephaestus/hephaestus-api/internal/svc"
	"jian-unified-system/hephaestus/hephaestus-api/internal/types"
	"jian-unified-system/hephaestus/hephaestus-rpc/income"
	"jian-unified-system/hephaestus/internal/transport"
)

type HealthLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHealthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HealthLogic {
	return &HealthLogic{ctx, svcCtx}
}
func (l *HealthLogic) Health(req *types.Empty) (*types.HealthResp, error) {
	input, e := access.Request[income.HealthRequest](l.ctx, req)
	if e != nil {
		return nil, e
	}
	v, e := l.svcCtx.Health.Health(l.ctx, input)
	if e != nil {
		return nil, e
	}
	return transport.Convert[types.HealthResp](v)
}
