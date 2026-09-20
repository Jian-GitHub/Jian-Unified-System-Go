// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth

import (
	"context"

	"jian-unified-system/hephaestus/hephaestus-api/internal/svc"
	"jian-unified-system/hephaestus/hephaestus-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuthCallbackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuthCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthCallbackLogic {
	return &AuthCallbackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AuthCallbackLogic) AuthCallback(req *types.AuthCallbackReq) error {
	return l.svcCtx.Access.Callback(l.ctx, req.Code, req.State, req.Error)
}
