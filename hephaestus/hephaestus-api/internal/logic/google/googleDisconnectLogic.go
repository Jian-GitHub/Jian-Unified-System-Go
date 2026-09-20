// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package google

import (
	"context"

	"jian-unified-system/hephaestus/hephaestus-api/internal/svc"
	"jian-unified-system/hephaestus/hephaestus-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GoogleDisconnectLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGoogleDisconnectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GoogleDisconnectLogic {
	return &GoogleDisconnectLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GoogleDisconnectLogic) GoogleDisconnect(req *types.Empty) (resp *types.GoogleConnection, err error) {
	v, e := l.svcCtx.Access.Google(l.ctx, "disconnect")
	if e != nil {
		return nil, e
	}
	return &types.GoogleConnection{Url: v.URL, Bound: v.Bound, Connected: v.Connected}, nil
}
