package ssologic

import (
	"context"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type IntrospectLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIntrospectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IntrospectLogic {
	return &IntrospectLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *IntrospectLogic) Introspect(in *apollo.SSOTokenReq) (*apollo.SSOSessionResp, error) {
	if e := authorizeRPC(l.ctx, l.svcCtx.Config.SSO.RPCSecret); e != nil {
		return nil, e
	}
	if in == nil {
		return nil, rpcError(identity.ErrInvalid)
	}
	v, e := l.svcCtx.SSO.Introspect(l.ctx, in.ClientId, in.ClientSecret, in.Token)
	return result(v, e)
}
