package thirdParty

import (
	"context"
	"encoding/json"
	"github.com/zeromicro/go-zero/core/errorx"
	"jian-unified-system/apollo/apollo-rpc/apollo"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetInfoLogic {
	return &GetInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetInfoLogic) GetInfo() (resp *types.GetInfoResp, err error) {
	id, err := l.ctx.Value("id").(json.Number).Int64()
	if err != nil {
		return &types.GetInfoResp{
			BaseResponse: types.BaseResponse{
				Code:    -1,
				Message: "id err",
			},
		}, errorx.Wrap(err, "token err")
	}

	rpcResp, err := l.svcCtx.ApolloThirdParty.GetInfo(l.ctx, &apollo.ThirdPartyGetInfoReq{
		UserId: id,
	})
	if err != nil {
		return nil, err
	}

	accounts := make([]types.ThirdPartyAccount, 0)
	if len(rpcResp.Accounts) > 0 {
		for _, account := range rpcResp.Accounts {
			accounts = append(accounts, types.ThirdPartyAccount{
				Id:       account.Id,
				Provider: account.Provider,
				Content:  account.Content,
			})
		}
	}

	return &types.GetInfoResp{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		GetInfoRespData: struct {
			Accounts []types.ThirdPartyAccount `json:"accounts"`
		}{
			Accounts: accounts,
		},
	}, nil
}
