// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package google

import (
	"context"
	"jian-unified-system/hephaestus/hephaestus-api/internal/access"
	"jian-unified-system/hephaestus/hephaestus-rpc/income"
	"jian-unified-system/hephaestus/internal/transport"

	"jian-unified-system/hephaestus/hephaestus-api/internal/svc"
	"jian-unified-system/hephaestus/hephaestus-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GoogleStageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGoogleStageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GoogleStageLogic {
	return &GoogleStageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GoogleStageLogic) GoogleStage(req *types.GoogleStageReq) (resp *types.Batch, err error) {
	in, e := access.Request[income.GoogleStageRequest](l.ctx, req)
	if e != nil {
		return nil, e
	}
	v, e := l.svcCtx.Imports.GoogleStage(l.ctx, in)
	if e != nil {
		return nil, e
	}
	return transport.Convert[types.Batch](v)
}
