package passkeys

import (
	"context"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PasskeysRegFinishLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPasskeysRegFinishLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PasskeysRegFinishLogic {
	return &PasskeysRegFinishLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PasskeysRegFinishLogic) PasskeysRegFinish(req *types.RegFinishReq) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
