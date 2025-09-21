package jquantumlogic

import (
	"context"

	"jian-unified-system/jquantum/jquantum-rpc/internal/svc"
	"jian-unified-system/jquantum/jquantum-rpc/jquantum"

	"github.com/zeromicro/go-zero/core/logx"
)

type ClusterInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewClusterInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClusterInfoLogic {
	return &ClusterInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ClusterInfo 集群信息
func (l *ClusterInfoLogic) ClusterInfo(in *jquantum.ClusterInfoReq) (*jquantum.ClusterInfoResp, error) {
	clusterResource, err := l.svcCtx.KubernetesDeployService.CollectClusterResource()
	if err != nil {
		return nil, err
	}

	return &jquantum.ClusterInfoResp{
		TotalCPU:  clusterResource.TotalCPU,
		TotalMem:  clusterResource.TotalMem,
		MaxQubits: clusterResource.MaxQubits,
		Nodes:     int64(len(clusterResource.Nodes)),
	}, nil
}
