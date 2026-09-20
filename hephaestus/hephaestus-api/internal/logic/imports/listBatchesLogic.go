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

type ListBatchesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListBatchesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListBatchesLogic {
	return &ListBatchesLogic{ctx, svcCtx}
}
func (l *ListBatchesLogic) ListBatches(req *types.Empty) (*types.BatchList, error) {
	input, e := access.Request[income.ListBatchesRequest](l.ctx, req)
	if e != nil {
		return nil, e
	}
	v, e := l.svcCtx.Imports.ListBatches(l.ctx, input)
	if e != nil {
		return nil, e
	}
	return transport.Convert[types.BatchList](v)
}
