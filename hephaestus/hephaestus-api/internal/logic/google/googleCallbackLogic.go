// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package google

import (
	"context"

	"jian-unified-system/hephaestus/hephaestus-api/internal/svc"
	"jian-unified-system/hephaestus/hephaestus-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GoogleCallbackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGoogleCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GoogleCallbackLogic {
	return &GoogleCallbackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GoogleCallbackLogic) GoogleCallback(req *types.GoogleCallbackReq) (resp *types.Empty, err error) {
	err = l.svcCtx.Access.GoogleCallback(l.ctx, req.State, req.Code, req.Error)
	return &types.Empty{}, err
}
