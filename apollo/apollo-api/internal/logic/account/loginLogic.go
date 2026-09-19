// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package account

import (
	"context"
	"strconv"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	p, token, err := l.svcCtx.Sessions.Login(l.ctx, req.Email, req.Password, req.CloudflareToken)
	if err != nil {
		return nil, err
	}
	return &types.LoginResp{BaseResponse: types.BaseResponse{Code: 200, Message: "success"}, Data: types.LoginData{
		Token: token, Id: strconv.FormatInt(p.ID, 10), Name: types.UserName{GivenName: p.GivenName, MiddleName: p.MiddleName, FamilyName: p.FamilyName},
		Avatar: p.Avatar, Locale: p.Locale, Language: p.Language, Birthday: types.Birthday{Year: p.BirthdayYear, Month: p.BirthdayMonth, Day: p.BirthdayDay},
	}}, nil
}
