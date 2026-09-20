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

type ListSourcesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListSourcesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSourcesLogic {
	return &ListSourcesLogic{ctx, svcCtx}
}
func (l *ListSourcesLogic) ListSources(req *types.Empty) (*types.SourceList, error) {
	input, e := access.Request[income.ListSourcesRequest](l.ctx, req)
	if e != nil {
		return nil, e
	}
	v, e := l.svcCtx.Imports.ListSources(l.ctx, input)
	if e != nil {
		return nil, e
	}
	return transport.Convert[types.SourceList](v)
}
