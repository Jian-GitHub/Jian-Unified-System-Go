package job

import (
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/errorx"
	"jian-unified-system/jquantum/jquantum-rpc/jquantum"

	"jian-unified-system/jquantum/jquantum-api/internal/svc"
	"jian-unified-system/jquantum/jquantum-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ClusterInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewClusterInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClusterInfoLogic {
	return &ClusterInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ClusterInfoLogic) ClusterInfo( /*req *types.ClusterInfoReq*/ ) (resp *types.ClusterInfoResp, err error) {
	clusterInfo, err := l.svcCtx.JQuantumClient.ClusterInfo(l.ctx, &jquantum.ClusterInfoReq{})
	if err != nil || clusterInfo == nil {
		return &types.ClusterInfoResp{
			BaseResponse: types.BaseResponse{
				Code:    -1,
				Message: "error",
			},
		}, errorx.Wrap(errors.New("JQuantumClient.ClusterInfo"), "System err")
	}

	return &types.ClusterInfoResp{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		TotalCPU:  clusterInfo.TotalCPU,
		TotalMem:  clusterInfo.TotalMem,
		MaxQubits: clusterInfo.MaxQubits,
		Nodes:     clusterInfo.Nodes,
	}, nil
}
