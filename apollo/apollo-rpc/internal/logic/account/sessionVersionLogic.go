package accountlogic

import (
	"context"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
	"jian-unified-system/apollo/apollo-rpc/internal/logic/response"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SessionVersionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSessionVersionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SessionVersionLogic {
	return &SessionVersionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SessionVersionLogic) SessionVersion(in *apollo.SessionVersionReq) (*apollo.SessionVersionResp, error) {
	if in == nil {
		return nil, response.Error(identity.ErrInvalid)
	}
	version, err := l.svcCtx.Account.SessionVersion(l.ctx, in.UserId)
	if err != nil {
		return nil, response.Error(err)
	}
	return &apollo.SessionVersionResp{AuthVersion: version}, nil
}
