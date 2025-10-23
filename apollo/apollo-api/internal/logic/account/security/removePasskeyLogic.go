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

type RemovePasskeyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRemovePasskeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemovePasskeyLogic {
	return &RemovePasskeyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RemovePasskeyLogic) RemovePasskey(req *types.RemovePasskeyReq) (resp *types.RemovePasskeyResp, err error) {
	// todo: add your logic here and delete this line
	id, err := l.ctx.Value("id").(json.Number).Int64()
	if err != nil {
		return &types.RemovePasskeyResp{
			BaseResponse: types.BaseResponse{
				Code:    -1,
				Message: "Id err",
			},
			RemovePasskeyData: struct {
				Ok bool `json:"ok"`
			}{
				Ok: false,
			},
		}, errorx.Wrap(errors.New("id"), "caller err")
	}

	passkeyResp, err := l.svcCtx.ApolloPasskeys.RemovePasskey(l.ctx, &apollo.RemovePasskeyReq{
		UserId:    id,
		PasskeyId: req.Id,
	})
	if err != nil {
		return &types.RemovePasskeyResp{
			BaseResponse: types.BaseResponse{
				Code:    -2,
				Message: "rpc err",
			},
			RemovePasskeyData: struct {
				Ok bool `json:"ok"`
			}{
				Ok: false,
			},
		}, err
	}

	return &types.RemovePasskeyResp{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		RemovePasskeyData: struct {
			Ok bool `json:"ok"`
		}{
			Ok: passkeyResp.Success,
		},
	}, nil
}
