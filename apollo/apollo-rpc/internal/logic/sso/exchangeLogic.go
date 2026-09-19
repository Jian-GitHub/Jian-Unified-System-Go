package ssologic

import (
	"context"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExchangeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewExchangeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExchangeLogic {
	return &ExchangeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ExchangeLogic) Exchange(in *apollo.SSOExchangeReq) (*apollo.SSOSessionResp, error) {
	if e := authorizeRPC(l.ctx, l.svcCtx.Config.SSO.RPCSecret); e != nil {
		return nil, e
	}
	if in == nil {
		return nil, rpcError(identity.ErrInvalid)
	}
	v, e := l.svcCtx.SSO.Exchange(l.ctx, in.ClientId, in.ClientSecret, in.Code, in.RedirectUri, in.CodeVerifier)
	return result(v, e)
}
