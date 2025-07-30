package account

import (
	"context"
	"github.com/zeromicro/go-zero/core/jsonx"

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

func (l *VerifyTokenLogic) VerifyToken(req *types.VerifyTokenReq) (resp *types.VerifyTokenResp, err error) {
	// todo: add your logic here and delete this line
	returns := make(map[string]any)
	returns["test"] = l.ctx.Value("test")

	dataJson, err := jsonx.MarshalToString(returns)
	if err != nil {
		return nil, err
	}

	return &types.VerifyTokenResp{
		BaseResponse: types.BaseResponse{
			Code:    0,
			Message: dataJson,
		},
		Data: struct {
			Test bool `json:"test"`
		}{
			Test: returns["test"].(bool),
		},
	}, nil
	//return &types.BaseResponse{
	//	Code:    202,
	//	Message: "success",
	//	Data:    dataJson,
	//}, nil
}
