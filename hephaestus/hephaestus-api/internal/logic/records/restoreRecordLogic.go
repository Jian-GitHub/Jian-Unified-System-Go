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

type RestoreRecordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRestoreRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RestoreRecordLogic {
	return &RestoreRecordLogic{ctx, svcCtx}
}
func (l *RestoreRecordLogic) RestoreRecord(req *types.IDReq) (*types.Record, error) {
	input, e := access.Request[income.RestoreRecordRequest](l.ctx, req)
	if e != nil {
		return nil, e
	}
	v, e := l.svcCtx.Records.RestoreRecord(l.ctx, input)
	if e != nil {
		return nil, e
	}
	return transport.Convert[types.Record](v)
}
