// Code scaffolded by goctl 1.9.2. Safe to edit: thin RPC adapter.
package imports

import (
	"context"
	"jian-unified-system/hephaestus/hephaestus-api/internal/access"
	"jian-unified-system/hephaestus/hephaestus-api/internal/svc"
	"jian-unified-system/hephaestus/hephaestus-api/internal/types"
	"jian-unified-system/hephaestus/hephaestus-rpc/income"
	"jian-unified-system/hephaestus/internal/transport"
)

type StageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewStageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StageLogic {
	return &StageLogic{ctx, svcCtx}
}
func (l *StageLogic) Stage(req *types.StageReq) (*types.Batch, error) {
	input, e := access.Request[income.StageRequest](l.ctx, req)
	if e != nil {
		return nil, e
	}
	v, e := l.svcCtx.Imports.Stage(l.ctx, input)
	if e != nil {
		return nil, e
	}
	return transport.Convert[types.Batch](v)
}
