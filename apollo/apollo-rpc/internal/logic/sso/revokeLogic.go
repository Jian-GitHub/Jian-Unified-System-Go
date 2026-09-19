package ssologic

import (
	"context"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RevokeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRevokeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RevokeLogic {
	return &RevokeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RevokeLogic) Revoke(in *apollo.SSOTokenReq) (*apollo.Empty, error) {
	if e := authorizeRPC(l.ctx, l.svcCtx.Config.SSO.RPCSecret); e != nil {
		return nil, e
	}
	if in == nil {
		return nil, rpcError(identity.ErrInvalid)
	}
	e := l.svcCtx.SSO.Revoke(l.ctx, in.ClientId, in.ClientSecret, in.Token)
	if e != nil {
		return nil, rpcError(e)
	}
	return &apollo.Empty{}, nil
}
