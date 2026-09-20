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

type GetRecordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRecordLogic {
	return &GetRecordLogic{ctx, svcCtx}
}
func (l *GetRecordLogic) GetRecord(req *types.IDReq) (*types.Record, error) {
	input, e := access.Request[income.GetRecordRequest](l.ctx, req)
	if e != nil {
		return nil, e
	}
	v, e := l.svcCtx.Records.GetRecord(l.ctx, input)
	if e != nil {
		return nil, e
	}
	return transport.Convert[types.Record](v)
}
