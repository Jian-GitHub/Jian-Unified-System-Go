package importslogic

import (
	"context"
	"jian-unified-system/hephaestus/hephaestus-rpc/internal/adapter"
	"jian-unified-system/hephaestus/internal/transport"

	"jian-unified-system/hephaestus/hephaestus-rpc/income"
	"jian-unified-system/hephaestus/hephaestus-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TrashBatchLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTrashBatchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TrashBatchLogic {
	return &TrashBatchLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *TrashBatchLogic) TrashBatch(in *income.TrashBatchRequest) (*income.Batch, error) {
	if in.Input == nil {
		in.Input = &income.TrashBatchReq{}
	}
	if e := adapter.Authorize(l.ctx, l.svcCtx.App, in.OwnerId, in.SessionToken); e != nil {
		return nil, transport.RPCError(e)
	}
	id, e := adapter.IntID(in.Input.Id)
	if e != nil {
		return nil, transport.RPCError(e)
	}
	v, e := l.svcCtx.App.TrashBatch(l.ctx, in.OwnerId, id, in.ExpectedVersion, in.IdempotencyKey, in.Input.Deleted)
	return transport.Result[income.Batch](v, e)
}
