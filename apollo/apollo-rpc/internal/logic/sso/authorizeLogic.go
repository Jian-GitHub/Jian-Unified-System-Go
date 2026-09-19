package ssologic

import (
	"context"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuthorizeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAuthorizeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthorizeLogic {
	return &AuthorizeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AuthorizeLogic) Authorize(in *apollo.SSOAuthorizeReq) (*apollo.SSOCodeResp, error) {
	if e := authorizeRPC(l.ctx, l.svcCtx.Config.SSO.RPCSecret); e != nil {
		return nil, e
	}
	if in == nil {
		return nil, rpcError(identity.ErrInvalid)
	}
	v, e := l.svcCtx.SSO.Authorize(l.ctx, in.UserId, in.AuthVersion, in.ClientId, in.RedirectUri, in.CodeChallenge)
	if e != nil {
		return nil, rpcError(e)
	}
	return &apollo.SSOCodeResp{Code: v}, nil
}
