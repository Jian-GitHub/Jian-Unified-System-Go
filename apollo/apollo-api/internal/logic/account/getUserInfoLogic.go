// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package account

import (
	"context"
	"jian-unified-system/apollo/apollo-api/internal/application"
	"jian-unified-system/apollo/apollo-api/internal/logic/response"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"strconv"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserInfoLogic) GetUserInfo(req *types.Empty) (resp *types.GetUserInfoResp, err error) {
	id, err := response.Subject(l.ctx)
	if err != nil {
		return nil, err
	}
	r, err := l.svcCtx.Account.UserInfo(l.ctx, &apollo.UserInfoReq{UserId: id})
	if err != nil {
		return nil, err
	}
	p := r.Profile
	if p == nil {
		return nil, application.ErrCredentials
	}
	return &types.GetUserInfoResp{BaseResponse: response.OK(), Data: types.UserInfoData{Id: strconv.FormatInt(p.UserId, 10), GivenName: p.GivenName, MiddleName: p.MiddleName, FamilyName: p.FamilyName, Avatar: p.Avatar, BirthdayYear: p.BirthdayYear, BirthdayMonth: p.BirthdayMonth, BirthdayDay: p.BirthdayDay, NotificationEmail: p.NotificationEmail, Locate: p.Locale, Language: p.Language, CreateTime: p.CreateTime, LastLoginTime: p.LastLoginTime}}, nil
}
