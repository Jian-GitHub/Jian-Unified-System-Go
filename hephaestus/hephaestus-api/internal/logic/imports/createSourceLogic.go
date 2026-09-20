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

type CreateSourceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateSourceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSourceLogic {
	return &CreateSourceLogic{ctx, svcCtx}
}
func (l *CreateSourceLogic) CreateSource(req *types.CreateSourceReq) (*types.Source, error) {
	input, e := access.Request[income.CreateSourceRequest](l.ctx, req)
	if e != nil {
		return nil, e
	}
	v, e := l.svcCtx.Imports.CreateSource(l.ctx, input)
	if e != nil {
		return nil, e
	}
	return transport.Convert[types.Source](v)
}
