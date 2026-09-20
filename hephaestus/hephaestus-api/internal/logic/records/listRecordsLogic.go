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

type ListRecordsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListRecordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRecordsLogic {
	return &ListRecordsLogic{ctx, svcCtx}
}
func (l *ListRecordsLogic) ListRecords(req *types.QueryReq) (*types.RecordList, error) {
	input, e := access.Request[income.ListRecordsRequest](l.ctx, req)
	if e != nil {
		return nil, e
	}
	v, e := l.svcCtx.Records.ListRecords(l.ctx, input)
	if e != nil {
		return nil, e
	}
	return transport.Convert[types.RecordList](v)
}
