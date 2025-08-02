package thirdParty

import (
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/errorx"
	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"
	apolloUtil "jian-unified-system/apollo/apollo-api/util"
	"jian-unified-system/jus-core/util/oauth2/redis"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
)

type ContinueLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewContinueLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ContinueLogic {
	return &ContinueLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ContinueLogic) Continue(req *types.ContinueReq, w http.ResponseWriter, r *http.Request) (resp *types.ContinueResp, err error) {
	// todo: add your logic here and delete this line
	if len(req.Provider) == 0 {
		return &types.ContinueResp{
			BaseResponse: types.BaseResponse{
				Code:    -1,
				Message: "no provider",
			},
		}, errorx.Wrap(errors.New("no provider"), "ThirdParty Continue Err")
	}

	// Save user id -> Redis.
	// Redis Key -> Third-Party.
	redisID := l.svcCtx.Snowflake.Generate().String()
	redis := redisUtil.NewContinueRedis(redisID)

	err = l.svcCtx.Redis.SetexCtx(l.ctx, redis.Key, redis.Data.String(), 300)
	if err != nil {
		return nil, err
	}

	url, err := apolloUtil.RedirectToOAuth2(l.svcCtx, req.Provider, redis.Key)
	if err != nil {
		return nil, err
	}

	return &types.ContinueResp{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		ContinueRespData: struct {
			Url string `json:"url"`
		}{
			Url: url,
		},
	}, nil
}
