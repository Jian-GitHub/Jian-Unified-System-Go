package thirdpartylogic

import (
	"context"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type BindLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBindLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindLogic {
	return &BindLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 绑定
func (l *BindLogic) Bind(in *apollo.ThirdPartyBindReq) (*apollo.Empty, error) {
	// todo: add your logic here and delete this line

	return &apollo.Empty{}, nil
}
