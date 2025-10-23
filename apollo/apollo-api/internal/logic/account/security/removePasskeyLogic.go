package security

import (
	"context"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemovePasskeyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRemovePasskeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemovePasskeyLogic {
	return &RemovePasskeyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RemovePasskeyLogic) RemovePasskey(req *types.RemovePasskeyReq) (resp *types.RemovePasskeyResp, err error) {
	// todo: add your logic here and delete this line

	return
}
