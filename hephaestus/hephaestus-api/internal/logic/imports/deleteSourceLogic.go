// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package imports

import (
	"context"
	"jian-unified-system/hephaestus/hephaestus-api/internal/access"
	"jian-unified-system/hephaestus/hephaestus-rpc/income"

	"jian-unified-system/hephaestus/hephaestus-api/internal/svc"
	"jian-unified-system/hephaestus/hephaestus-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteSourceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteSourceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSourceLogic {
	return &DeleteSourceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteSourceLogic) DeleteSource(req *types.IDReq) (resp *types.Empty, err error) {
	in, e := access.Request[income.DeleteSourceRequest](l.ctx, req)
	if e != nil {
		return nil, e
	}
	_, e = l.svcCtx.Imports.DeleteSource(l.ctx, in)
	return &types.Empty{}, e
}
