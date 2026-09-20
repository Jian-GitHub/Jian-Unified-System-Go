// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth

import (
	"context"

	"jian-unified-system/hephaestus/hephaestus-api/internal/svc"
	"jian-unified-system/hephaestus/hephaestus-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuthStartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuthStartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthStartLogic {
	return &AuthStartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AuthStartLogic) AuthStart(req *types.AuthStartReq) error {
	return l.svcCtx.Access.Start(l.ctx, req.ReturnTo)
}
