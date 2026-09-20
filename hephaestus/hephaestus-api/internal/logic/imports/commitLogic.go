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

type CommitLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCommitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CommitLogic {
	return &CommitLogic{ctx, svcCtx}
}
func (l *CommitLogic) Commit(req *types.CommitReq) (*types.Batch, error) {
	input, e := access.Request[income.CommitRequest](l.ctx, req)
	if e != nil {
		return nil, e
	}
	v, e := l.svcCtx.Imports.Commit(l.ctx, input)
	if e != nil {
		return nil, e
	}
	return transport.Convert[types.Batch](v)
}
