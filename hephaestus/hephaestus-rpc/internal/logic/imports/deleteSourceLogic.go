package importslogic

import (
	"context"
	"jian-unified-system/hephaestus/hephaestus-rpc/internal/adapter"
	d "jian-unified-system/hephaestus/hephaestus-rpc/internal/domain"
	"jian-unified-system/hephaestus/internal/transport"

	"jian-unified-system/hephaestus/hephaestus-rpc/income"
	"jian-unified-system/hephaestus/hephaestus-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteSourceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteSourceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSourceLogic {
	return &DeleteSourceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteSourceLogic) DeleteSource(in *income.DeleteSourceRequest) (*income.Empty, error) {
	if in == nil || in.Input == nil {
		return nil, transport.RPCError(d.Invalid("source id required"))
	}
	if e := adapter.Authorize(l.ctx, l.svcCtx.App, in.OwnerId, in.SessionToken); e != nil {
		return nil, transport.RPCError(e)
	}
	id, e := d.ID(in.Input.Id)
	if e != nil {
		return nil, transport.RPCError(e)
	}
	e = l.svcCtx.App.DeleteSource(l.ctx, in.OwnerId, id, in.IdempotencyKey)
	return &income.Empty{}, transport.RPCError(e)
}
