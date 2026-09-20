// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package imports

import (
	"context"
	"jian-unified-system/hephaestus/hephaestus-api/internal/access"
	"jian-unified-system/hephaestus/hephaestus-rpc/income"
	"jian-unified-system/hephaestus/internal/transport"

	"jian-unified-system/hephaestus/hephaestus-api/internal/svc"
	"jian-unified-system/hephaestus/hephaestus-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReadSheetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReadSheetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReadSheetLogic {
	return &ReadSheetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReadSheetLogic) ReadSheet(req *types.SheetReq) (resp *types.SheetData, err error) {
	in, e := access.Request[income.ReadSheetRequest](l.ctx, req)
	if e != nil {
		return nil, e
	}
	v, e := l.svcCtx.Imports.ReadSheet(l.ctx, in)
	if e != nil {
		return nil, e
	}
	return transport.Convert[types.SheetData](v)
}
