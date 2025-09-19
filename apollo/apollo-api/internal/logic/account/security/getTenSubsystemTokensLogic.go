package security

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/zeromicro/go-zero/core/errorx"
	"jian-unified-system/apollo/apollo-rpc/apollo"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenSubsystemTokensLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTenSubsystemTokensLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenSubsystemTokensLogic {
	return &GetTenSubsystemTokensLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTenSubsystemTokensLogic) GetTenSubsystemTokens(req *types.GetTenSubsystemTokensReq) (resp *types.GetTenSubsystemTokensResp, err error) {
	id, err := l.ctx.Value("id").(json.Number).Int64()
	if err != nil {
		return &types.GetTenSubsystemTokensResp{
			BaseResponse: types.BaseResponse{
				Code:    -2,
				Message: "Id err",
			},
		}, errorx.Wrap(errors.New("id"), "caller err")
	}

	tokensResp, err := l.svcCtx.ApolloSecurity.FindTenSubsystemTokens(l.ctx, &apollo.FindTenSubsystemTokensReq{
		UserId: id,
		Page:   req.Page,
	})
	if err != nil {
		return nil, err
	}

	tokens := make([]types.SubsystemToken, 0)
	if len(tokensResp.Tokens) != 0 {
		for _, token := range tokensResp.Tokens {
			tokens = append(tokens, types.SubsystemToken{
				Id:    token.Id,
				Value: token.Value,
				Name:  token.Name,
				Date: types.RespnseDate{
					Year:  token.Year,
					Month: token.Month,
					Day:   token.Day,
				},
			})
		}
	}

	return &types.GetTenSubsystemTokensResp{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		GetAllSubsystemTokensData: struct {
			Tokens []types.SubsystemToken `json:"tokens"`
		}{
			Tokens: tokens,
		},
	}, nil
}
