// Code scaffolded by goctl 1.9.2. Safe to edit: thin RPC adapter.
package reporting

import (
	"context"
	"jian-unified-system/hephaestus/hephaestus-api/internal/access"
	"jian-unified-system/hephaestus/hephaestus-api/internal/svc"
	"jian-unified-system/hephaestus/hephaestus-api/internal/types"
	"jian-unified-system/hephaestus/hephaestus-rpc/income"
	"jian-unified-system/hephaestus/internal/transport"
)

type ExportLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewExportLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExportLogic {
	return &ExportLogic{ctx, svcCtx}
}
func (l *ExportLogic) Export(req *types.QueryReq) (*types.ExportResp, error) {
	input, e := access.Request[income.ExportRequest](l.ctx, req)
	if e != nil {
		return nil, e
	}
	v, e := l.svcCtx.Reporting.Export(l.ctx, input)
	if e != nil {
		return nil, e
	}
	return transport.Convert[types.ExportResp](v)
}
