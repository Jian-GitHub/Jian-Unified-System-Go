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

type ListRowsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListRowsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRowsLogic {
	return &ListRowsLogic{ctx, svcCtx}
}
func (l *ListRowsLogic) ListRows(req *types.BatchRowsReq) (*types.BatchRows, error) {
	input, e := access.Request[income.ListRowsRequest](l.ctx, req)
	if e != nil {
		return nil, e
	}
	v, e := l.svcCtx.Imports.ListRows(l.ctx, input)
	if e != nil {
		return nil, e
	}
	return transport.Convert[types.BatchRows](v)
}
