package account

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/errorx"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/jus-core/util"

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
	// todo: add your logic here and delete this line
	// Check params
	if len(req.Email) == 0 || len(req.Password) == 0 {
		return &types.LoginResp{
			BaseResponse: types.BaseResponse{
				Code:    -1,
				Message: "params",
			},
			LoginData: struct {
				Token string `json:"token"`
			}{
				Token: "",
			},
		}, errorx.Wrap(err, "params")
	}

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
			Message: "params",
		},
		LoginData: struct {
			Token string `json:"token"`
		}{
			Token: token,
		},
	}, nil
}
