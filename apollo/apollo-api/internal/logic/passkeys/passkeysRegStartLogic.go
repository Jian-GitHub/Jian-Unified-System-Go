package passkeys

import (
	"context"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PasskeysRegStartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPasskeysRegStartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PasskeysRegStartLogic {
	return &PasskeysRegStartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PasskeysRegStartLogic) PasskeysRegStart(req *types.RegStartReq) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
