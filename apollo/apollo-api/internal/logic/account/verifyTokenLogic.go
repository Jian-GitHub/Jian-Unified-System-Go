package account

import (
	"context"
	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type VerifyTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVerifyTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyTokenLogic {
	return &VerifyTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VerifyTokenLogic) VerifyToken() (resp *types.VerifyTokenResp, err error) {
	//returns := make(map[string]any)
	//returns["id"] = l.ctx.Value("id")
	//
	//dataJson, err := jsonx.MarshalToString(returns)
	//if err != nil {
	//	return nil, err
	//}

	return &types.VerifyTokenResp{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		Data: struct {
			Ok bool `json:"ok"`
		}{
			Ok: true,
		},
	}, nil
}
