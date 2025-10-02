package thirdParty

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/errorx"
	apolloUtil "jian-unified-system/apollo/apollo-api/util"
	redisUtil "jian-unified-system/jus-core/util/oauth2/redis"
	"strconv"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BindLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBindLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindLogic {
	return &BindLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BindLogic) Bind(req *types.BindReq) (resp *types.BindResp, err error) {
	// todo: add your logic here and delete this line
	if len(req.Provider) == 0 {
		return &types.BindResp{
			BaseResponse: types.BaseResponse{
				Code:    -1,
				Message: "no provider",
			},
		}, errorx.Wrap(errors.New("no provider"), "ThirdParty Continue Err")
	}
	id, err := l.ctx.Value("id").(json.Number).Int64()
	if err != nil {
		return &types.BindResp{
			BaseResponse: types.BaseResponse{
				Code:    -1,
				Message: "no provider",
			},
		}, errorx.Wrap(err, "token err")
	}
	fmt.Println(id)

	idStr := strconv.FormatInt(id, 10)
	redis := redisUtil.NewBindRedis(idStr)
	//redisKey := "apollo:thirdParty:bind:" + hex.EncodeToString([]byte(strconv.FormatInt(id, 10)))

	err = l.svcCtx.Redis.SetexCtx(l.ctx, redis.Key, redis.Data.String(), 300)
	if err != nil {
		return nil, err
	}

	url, err := apolloUtil.RedirectToOAuth2(l.svcCtx, req.Provider, redis.Key)
	if err != nil {
		return nil, err
	}

	return &types.BindResp{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		BindRespData: struct {
			Url string `json:"url"`
		}{
			Url: url,
		},
	}, nil
}
