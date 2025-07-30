package passkeys

import (
	"context"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PasskeysLoginFinishLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPasskeysLoginFinishLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PasskeysLoginFinishLogic {
	return &PasskeysLoginFinishLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PasskeysLoginFinishLogic) PasskeysLoginFinish(req *types.LoginFinishReq) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
