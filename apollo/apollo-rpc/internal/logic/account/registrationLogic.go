package accountlogic

import (
	"context"
	accountapp "jian-unified-system/apollo/apollo-rpc/internal/application/account"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegistrationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegistrationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegistrationLogic {
	return &RegistrationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegistrationLogic) Registration(in *apollo.RegistrationReq) (*apollo.Empty, error) {
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid input")
	}
	err := l.svcCtx.Account.Register(l.ctx, accountapp.Registration{ID: in.UserId, Email: in.Email, Password: in.Password, Locale: in.Locate, Language: in.Language})
	if err != nil {
		return nil, transportError(err)
	}
	return &apollo.Empty{}, nil
}
