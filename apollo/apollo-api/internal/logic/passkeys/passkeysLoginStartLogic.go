package passkeys

import (
	"context"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PasskeysLoginStartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPasskeysLoginStartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PasskeysLoginStartLogic {
	return &PasskeysLoginStartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PasskeysLoginStartLogic) PasskeysLoginStart(req *types.Empty) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
