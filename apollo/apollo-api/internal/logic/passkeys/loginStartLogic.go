package passkeys

import (
	"context"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginStartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginStartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginStartLogic {
	return &LoginStartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginStartLogic) LoginStart(req *types.Empty) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
