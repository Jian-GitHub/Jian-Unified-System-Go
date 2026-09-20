package importslogic

import (
	"context"
	"jian-unified-system/hephaestus/hephaestus-rpc/internal/adapter"
	d "jian-unified-system/hephaestus/hephaestus-rpc/internal/domain"
	"jian-unified-system/hephaestus/internal/transport"
	"strconv"

	"jian-unified-system/hephaestus/hephaestus-rpc/income"
	"jian-unified-system/hephaestus/hephaestus-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GoogleStageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGoogleStageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GoogleStageLogic {
	return &GoogleStageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GoogleStageLogic) GoogleStage(in *income.GoogleStageRequest) (*income.Batch, error) {
	if in == nil || in.Input == nil {
		return nil, transport.RPCError(d.Invalid("missing Google import input"))
	}
	if e := adapter.Authorize(l.ctx, l.svcCtx.App, in.OwnerId, in.SessionToken); e != nil {
		return nil, transport.RPCError(e)
	}
	source, e := strconv.ParseInt(in.Input.SourceId, 10, 64)
	if e != nil || source <= 0 {
		return nil, transport.RPCError(d.Invalid("invalid source"))
	}
	v, e := l.svcCtx.App.StageGoogle(l.ctx, in.OwnerId, source, in.IdempotencyKey, in.SessionToken, in.Input.Spreadsheet, in.Input.Range)
	return transport.Result[income.Batch](v, e)
}
