package passkeyslogic

import (
	"context"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemovePasskeyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRemovePasskeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemovePasskeyLogic {
	return &RemovePasskeyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RemovePasskey 移除 Passkey
func (l *RemovePasskeyLogic) RemovePasskey(in *apollo.RemovePasskeyReq) (*apollo.RemovePasskeyResp, error) {
	// todo: add your logic here and delete this line

	return &apollo.RemovePasskeyResp{}, nil
}
