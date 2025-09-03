package job

import (
	"context"
	"encoding/json"
	"github.com/zeromicro/go-zero/core/errorx"
	"jian-unified-system/jquantum/jquantum-rpc/jquantum"

	"jian-unified-system/jquantum/jquantum-api/internal/svc"
	"jian-unified-system/jquantum/jquantum-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RetrieveResultLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRetrieveResultLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetrieveResultLogic {
	return &RetrieveResultLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RetrieveResultLogic) RetrieveResult(req *types.RetrieveResultReq) (resp *types.RetrieveResultResp, err error) {
	// todo: add your logic here and delete this line
	id, err := l.ctx.Value("id").(json.Number).Int64()
	if err != nil {
		return &types.RetrieveResultResp{
			BaseResponse: types.BaseResponse{
				Code:    -1,
				Message: "No user id",
			},
		}, errorx.Wrap(err, "No user id")
	}

	result, err := l.svcCtx.JQuantumClient.RetrieveResult(l.ctx, &jquantum.RetrieveResultReq{
		UserId: id,
		JobId:  req.JobID,
	})
	if err != nil {
		return &types.RetrieveResultResp{
			BaseResponse: types.BaseResponse{
				Code:    2,
				Message: "rpc err",
			},
		}, errorx.Wrap(err, "rpc err")
	}

	return &types.RetrieveResultResp{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		Result: result.Result,
	}, nil
}
