package ssologic

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LogoutLogic) Logout(in *apollo.SSOLogoutReq) (*apollo.Empty, error) {
	if e := authorizeRPC(l.ctx, l.svcCtx.Config.SSO.RPCSecret); e != nil {
		return nil, e
	}
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid input")
	}
	if e := l.svcCtx.SSO.Logout(l.ctx, in.UserId); e != nil {
		return nil, rpcError(e)
	}
	return &apollo.Empty{}, nil
}
