package account

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/errorx"
	apolloUtil "jian-unified-system/apollo/apollo-api/util"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/jus-core/util"
	"net/http"
	"strings"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegLogic {
	return &RegLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegLogic) Reg(req *types.RegReq, r *http.Request) (resp *types.RegResp, err error) {
	// todo: add your logic here and delete this line

	//fmt.Println("ip: ", info.Country)
	//fmt.Println("Country: ", ip)
	//fmt.Println("City: ", info.City)
	//fmt.Println("Region: ", info.Region)
	//fmt.Println("IsoCode: ", info.IsoCode)
	//fmt.Println("info: ", info)

	//str, err := l.svcCtx.GeoService.GetLocalizedRegionName(info.IsoCode, "en-NZ")
	//if err != nil {
	//	fmt.Println("err:", err.Error())
	//	return nil, err
	//}
	//fmt.Println("getLocalizedRegionName, ", str)

	//return nil, nil
	// Check params
	if len(req.Email) == 0 || len(req.Password) == 0 || len(req.ConfirmPassword) == 0 || len(req.Language) == 0 || strings.Compare(req.Password, req.ConfirmPassword) != 0 {
		fmt.Println(len(req.Email) == 0)
		fmt.Println(len(req.Password) == 0)
		fmt.Println(len(req.ConfirmPassword) == 0)
		fmt.Println(len(req.Language) == 0)
		fmt.Println(strings.Compare(req.Password, req.ConfirmPassword) != 0)
		return &types.RegResp{
			BaseResponse: types.BaseResponse{
				Code:    -1,
				Message: "params",
			},
		}, errorx.Wrap(err, "params")
	}

	// Check Email format
	if !util.IsEmail(req.Email) {
		return &types.RegResp{
			BaseResponse: types.BaseResponse{
				Code:    -3,
				Message: "params: email",
			},
		}, errorx.Wrap(err, "Eamil")
	}

	// Check Pwd format
	if !util.IsStrongPassword(req.Password) {
		return &types.RegResp{
			BaseResponse: types.BaseResponse{
				Code:    -4,
				Message: "params: pwd",
			},
		}, errorx.Wrap(err, "Password")
	}
	// Check Finished

	// Get Locate
	//ip := apolloUtil.GetRealIP(r)
	//locate := "CN"
	//info, err := l.svcCtx.GeoService.Lookup(ip)
	//if err == nil && info != nil {
	//	locate = info.IsoCode
	//}
	locate := apolloUtil.GetLocate(r, l.svcCtx.GeoService.Lookup)

	// Generate a new User -> rpc saves new user
	id := l.svcCtx.Snowflake.Generate().Int64()
	_, err = l.svcCtx.ApolloAccount.Registration(l.ctx,
		&apollo.RegistrationReq{
			UserId:   id,
			Email:    req.Email,
			Password: req.Password,
			Locate:   locate,
			Language: req.Language,
		})
	if err != nil {
		fmt.Println(err.Error())
		return nil, errorx.Wrap(err, "rpc error: reg user")
	}

	// All done -> Generate JWT
	args := make(map[string]interface{})
	args["id"] = id
	token, err := util.GenToken(l.svcCtx.Config.Auth.AccessSecret, l.svcCtx.Config.Auth.AccessExpire, args)
	if err != nil {
		return &types.RegResp{
			BaseResponse: types.BaseResponse{
				Code:    -2,
				Message: "token err",
			},
			RegRespData: struct {
				Token string `json:"token"`
			}{Token: ""},
		}, errorx.Wrap(err, "token err")
	}

	return &types.RegResp{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		RegRespData: struct {
			Token string `json:"token"`
		}{Token: token},
	}, nil
}
