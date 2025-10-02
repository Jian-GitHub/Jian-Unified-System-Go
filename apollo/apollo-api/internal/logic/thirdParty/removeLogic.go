package thirdParty

import (
	"context"
	"encoding/json"
	"github.com/zeromicro/go-zero/core/errorx"
	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"
	"jian-unified-system/apollo/apollo-rpc/apollo"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRemoveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveLogic {
	return &RemoveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RemoveLogic) Remove(req *types.RemoveReq) (resp *types.RemoveResp, err error) {
	id, err := l.ctx.Value("id").(json.Number).Int64()
	if err != nil {
		return &types.RemoveResp{
			BaseResponse: types.BaseResponse{
				Code:    -1,
				Message: "id err",
			},
		}, errorx.Wrap(err, "token err")
	}

	_, err = l.svcCtx.ApolloThirdParty.Remove(l.ctx, &apollo.ThirdPartyRemoveReq{
		UserId:       id,
		ThirdPartyId: req.ThirdPartyId,
	})
	if err != nil {
		return &types.RemoveResp{
			BaseResponse: types.BaseResponse{
				Code:    -3,
				Message: "rpc err",
			},
		}, errorx.Wrap(err, "rpc err")
	}

	return &types.RemoveResp{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		RemoveRespData: struct {
			Ok bool `json:"ok"`
		}{
			true,
		},
	}, nil
}
