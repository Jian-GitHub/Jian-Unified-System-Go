package account

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/errorx"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/jus-core/util"
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

func (l *LoginLogic) Login(req *types.LoginReq /*, r *http.Request*/) (resp *types.LoginResp, err error) {
	// Check params
	if len(req.Email) == 0 || len(req.Password) == 0 {
		return &types.LoginResp{
			BaseResponse: types.BaseResponse{
				Code:    -1,
				Message: "params",
			},
		}, errorx.Wrap(err, "params")
	}
	//println(apolloUtil.GetRealIP(r))
	//println(apolloUtil.GetLocate(r, l.svcCtx.GeoService.Lookup))

	// Login
	loginResp, err := l.svcCtx.ApolloAccount.Login(l.ctx, &apollo.LoginReq{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		fmt.Println(err.Error())
		return nil, errorx.Wrap(err, "login fail")
	}

	// All done -> Generate JWT
	args := make(map[string]interface{})
	args["id"] = loginResp.UserId
	token, err := util.GenToken(l.svcCtx.Config.Auth.AccessSecret, l.svcCtx.Config.Auth.AccessExpire, args)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	return &types.LoginResp{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		LoginData: struct {
			Token string `json:"token"`
			Id    string `json:"id"`
			Name  struct {
				GivenName  string `json:"givenName"`
				MiddleName string `json:"middleName"`
				FamilyName string `json:"familyName"`
			} `json:"name"`
			Avatar   string `json:"avatar"`
			Locale   string `json:"locale"`
			Language string `json:"language"`
			Birthday struct {
				Year  int64 `json:"year"`
				Month int64 `json:"month"`
				Day   int64 `json:"day"`
			} `json:"birthday"`
		}{
			Token: token,
			Id:    strconv.FormatInt(loginResp.UserId, 10),
			Name: struct {
				GivenName  string `json:"givenName"`
				MiddleName string `json:"middleName"`
				FamilyName string `json:"familyName"`
			}{
				GivenName:  loginResp.GivenName,
				MiddleName: loginResp.MiddleName,
				FamilyName: loginResp.FamilyName,
			},
			Avatar:   loginResp.Avatar,
			Locale:   loginResp.Locale,
			Language: loginResp.Language,
			Birthday: struct {
				Year  int64 `json:"year"`
				Month int64 `json:"month"`
				Day   int64 `json:"day"`
			}{
				Year:  loginResp.BirthdayYear,
				Month: loginResp.BirthdayMonth,
				Day:   loginResp.BirthdayDay,
			},
		},
	}, nil
}
