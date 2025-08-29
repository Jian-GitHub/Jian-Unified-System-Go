package joblogic

import (
	"context"

	"jian-unified-system/jquantum/jquantum-rpc/internal/svc"
	"jian-unified-system/jquantum/jquantum-rpc/jquantum"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubmitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitLogic {
	return &SubmitLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 注册
func (l *SubmitLogic) Submit(in *jquantum.Empty) (*jquantum.Empty, error) {
	// todo: add your logic here and delete this line

	return &jquantum.Empty{}, nil
}
