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

type CreateRecordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRecordLogic {
	return &CreateRecordLogic{ctx, svcCtx}
}
func (l *CreateRecordLogic) CreateRecord(req *types.CreateRecordReq) (*types.Record, error) {
	input, e := access.Request[income.CreateRecordRequest](l.ctx, req)
	if e != nil {
		return nil, e
	}
	v, e := l.svcCtx.Records.CreateRecord(l.ctx, input)
	if e != nil {
		return nil, e
	}
	return transport.Convert[types.Record](v)
}
