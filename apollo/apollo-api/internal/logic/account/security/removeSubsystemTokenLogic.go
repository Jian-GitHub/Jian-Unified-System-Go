package security

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/zeromicro/go-zero/core/errorx"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"strconv"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveSubsystemTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRemoveSubsystemTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveSubsystemTokenLogic {
	return &RemoveSubsystemTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RemoveSubsystemTokenLogic) RemoveSubsystemToken(req *types.RemoveSubsystemTokenReq) (resp *types.RemoveSubsystemTokenResp, err error) {
	// todo: add your logic here and delete this line
	id, err := l.ctx.Value("id").(json.Number).Int64()
	if err != nil {
		return &types.RemoveSubsystemTokenResp{
			BaseResponse: types.BaseResponse{
				Code:    -1,
				Message: "Id err",
			},
		}, errorx.Wrap(errors.New("id"), "caller err")
	}

	tokenId, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return &types.RemoveSubsystemTokenResp{
			BaseResponse: types.BaseResponse{
				Code:    -2,
				Message: "tokenId err",
			},
		}, errorx.Wrap(errors.New("tokenId"), "caller err")
	}

	removeSubsystemTokenResp, err := l.svcCtx.ApolloSecurity.RemoveSubsystemToken(l.ctx, &apollo.RemoveSubsystemTokenReq{
		UserId:  id,
		TokenId: tokenId,
	})
	if err != nil {
		return &types.RemoveSubsystemTokenResp{
			BaseResponse: types.BaseResponse{
				Code:    -3,
				Message: "rpc error",
			},
		}, err
	}
	return &types.RemoveSubsystemTokenResp{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		RemoveSubsystemTokenData: struct {
			Ok bool `json:"ok"`
		}{
			Ok: removeSubsystemTokenResp.Validated,
		},
	}, nil
}
