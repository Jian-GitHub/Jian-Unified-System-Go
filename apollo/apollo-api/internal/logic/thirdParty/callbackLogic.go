package thirdParty

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/errorx"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/jus-core/util"
	"strings"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CallbackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CallbackLogic {
	return &CallbackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CallbackLogic) Callback(req *types.CallbackReq) (resp *types.CallbackResp, err error) {
	// todo: add your logic here and delete this line
	if len(strings.TrimSpace(req.Provider)) == 0 || len(req.State) == 0 || len(req.Code) == 0 {
		return nil, errorx.Wrap(errors.New("param is empty"), "ThirdParty Err")
	}
	// Redis --state--> redis data string
	redisDataJson, err := l.svcCtx.Redis.GetCtx(l.ctx, req.State)
	if err != nil {
		return nil, err
	}
	// redis data string -> RedisData
	//fmt.Println(redisDataJson)
	//var redisData *redisUtil.RedisData
	//err = json.Unmarshal([]byte(redisDataJson), &redisData)
	//if err != nil {
	//	return nil, errorx.Wrap(errors.New("redis state fail"), "ThirdParty Err")
	//}
	//var id int64 = 0
	//switch redisData.Flag {
	//case redisUtil.BindFlag:
	//	id, err = strconv.ParseInt(*redisData.Id, 10, 64)
	//	if err != nil {
	//		return nil, err
	//	}
	//	//case redisUtil.ContinueFlag:
	//}
	// id string -> id int64
	//id, err := strconv.ParseInt(idStr, 10, 64)
	//if err != nil {
	//	return nil, err
	//}

	// Provider -> OAuth2 Config
	cfg, ok := l.svcCtx.OauthProviders[req.Provider]
	if !ok {
		return &types.CallbackResp{
			BaseResponse: types.BaseResponse{
				Code:    -2,
				Message: "no such provider",
			},
		}, errorx.Wrap(errors.New("no provider"), "ThirdParty Continue Err")
	}

	// code -> token
	token, err := cfg.Exchange(l.ctx, req.Code)
	if err != nil {
		return &types.CallbackResp{
			BaseResponse: types.BaseResponse{
				Code:    -3,
				Message: "no such provider",
			},
		}, errorx.Wrap(errors.New("exchange token fail"), "ThirdParty Callback Err")
	}

	tokenBytes, err := json.Marshal(token)
	if err != nil {
		return nil, err
	}

	// grpc
	result, err := l.svcCtx.ApolloThirdParty.HandleCallback(l.ctx, &apollo.ThirdPartyContinueReq{
		//Id:            id,
		Provider:      req.Provider,
		Token:         tokenBytes,
		RedisDataJson: redisDataJson,
	})
	if err != nil {
		return nil, err
	}

	// All done -> Generate JWT
	args := make(map[string]interface{})
	args["id"] = result.UserId
	apolloToken, err := util.GenToken(l.svcCtx.Config.Auth.AccessSecret, l.svcCtx.Config.Auth.AccessExpire, args)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return &types.CallbackResp{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		CallbackRespData: struct {
			Token string `json:"token"`
		}{
			Token: apolloToken,
		},
	}, nil
}
