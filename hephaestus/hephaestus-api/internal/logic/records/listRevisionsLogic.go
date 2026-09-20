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

type ListRevisionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListRevisionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRevisionsLogic {
	return &ListRevisionsLogic{ctx, svcCtx}
}
func (l *ListRevisionsLogic) ListRevisions(req *types.IDReq) (*types.RevisionList, error) {
	input, e := access.Request[income.ListRevisionsRequest](l.ctx, req)
	if e != nil {
		return nil, e
	}
	v, e := l.svcCtx.Records.ListRevisions(l.ctx, input)
	if e != nil {
		return nil, e
	}
	return transport.Convert[types.RevisionList](v)
}
