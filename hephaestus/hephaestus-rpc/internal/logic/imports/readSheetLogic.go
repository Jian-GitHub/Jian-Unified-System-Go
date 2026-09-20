package importslogic

import (
	"context"
	"jian-unified-system/hephaestus/hephaestus-rpc/internal/adapter"
	"jian-unified-system/hephaestus/internal/transport"

	"jian-unified-system/hephaestus/hephaestus-rpc/income"
	"jian-unified-system/hephaestus/hephaestus-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReadSheetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReadSheetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReadSheetLogic {
	return &ReadSheetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ReadSheetLogic) ReadSheet(in *income.ReadSheetRequest) (*income.SheetData, error) {
	if in.Input == nil {
		in.Input = &income.SheetReq{}
	}
	if e := adapter.Authorize(l.ctx, l.svcCtx.App, in.OwnerId, in.SessionToken); e != nil {
		return nil, transport.RPCError(e)
	}
	id, e := adapter.IntID(in.Input.Id)
	if e != nil {
		return nil, transport.RPCError(e)
	}
	v, e := l.svcCtx.App.ReadSheet(l.ctx, in.OwnerId, id, in.Input.Sheet, in.Input.Cursor)
	return transport.Result[income.SheetData](v, e)
}
