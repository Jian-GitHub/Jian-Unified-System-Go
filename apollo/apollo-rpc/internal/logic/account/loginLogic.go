package accountlogic

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *apollo.LoginReq) (*apollo.LoginResp, error) {
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid input")
	}
	p, err := l.svcCtx.Account.Login(l.ctx, in.Email, in.Password)
	if err != nil {
		return nil, transportError(err)
	}
	return &apollo.LoginResp{UserId: p.ID(), GivenName: p.GivenName(), MiddleName: p.MiddleName(), FamilyName: p.FamilyName(), Avatar: p.Avatar(), Locale: p.Locale(), Language: p.Language(), BirthdayYear: p.BirthdayYear(), BirthdayMonth: p.BirthdayMonth(), BirthdayDay: p.BirthdayDay(), AuthVersion: p.AuthVersion()}, nil
}
