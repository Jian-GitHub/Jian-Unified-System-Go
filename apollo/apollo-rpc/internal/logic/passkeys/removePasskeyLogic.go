package passkeyslogic

import (
	"context"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
	"jian-unified-system/apollo/apollo-rpc/internal/logic/response"

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

func (l *RemovePasskeyLogic) RemovePasskey(in *apollo.RemovePasskeyReq) (*apollo.RemovePasskeyResp, error) {
	if in == nil {
		return nil, response.Error(identity.ErrInvalid)
	}
	if err := l.svcCtx.Passkey.Remove(l.ctx, in.UserId, in.PasskeyId); err != nil {
		return nil, response.Error(err)
	}
	return &apollo.RemovePasskeyResp{Success: true}, nil
}
