package accountlogic

import (
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/errorx"
	"jian-unified-system/jus-core/util"

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

// Login 登录
func (l *LoginLogic) Login(in *apollo.LoginReq) (*apollo.LoginResp, error) {
	email := util.HashSHA512(in.Email, in.Email)
	user, err := l.svcCtx.UserModel.FindOneByEmail(l.ctx, email)
	if err != nil {
		return nil, errorx.Wrap(errors.New("no user found"), "login sql err")
	}
	//fmt.Println(user)
	if user == nil {
		return nil, errorx.Wrap(errors.New("null user"), "login sql err")
	}
	if !util.VerifyPasswordBcrypt(in.Password, user.Password) {
		return nil, errorx.Wrap(errors.New("password Error"), "Login Error")
	}

	return &apollo.LoginResp{
		UserId:     user.Id,
		GivenName:  user.GivenName,
		MiddleName: user.MiddleName,
		FamilyName: user.FamilyName,
		Avatar:     user.Avatar.String,
		Locale:     user.Locate,
		Language:   user.Language,
	}, nil
}
