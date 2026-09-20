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

type SeriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSeriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SeriesLogic {
	return &SeriesLogic{ctx, svcCtx}
}
func (l *SeriesLogic) Series(req *types.QueryReq) (*types.PeriodList, error) {
	input, e := access.Request[income.SeriesRequest](l.ctx, req)
	if e != nil {
		return nil, e
	}
	v, e := l.svcCtx.Reporting.Series(l.ctx, input)
	if e != nil {
		return nil, e
	}
	return transport.Convert[types.PeriodList](v)
}
