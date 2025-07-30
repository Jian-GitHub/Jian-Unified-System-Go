package passkeyslogic

import (
	"context"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type StartLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewStartLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StartLoginLogic {
	return &StartLoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 登录
func (l *StartLoginLogic) StartLogin(in *apollo.Empty) (*apollo.PasskeysStartLoginResp, error) {
	// todo: add your logic here and delete this line

	return &apollo.PasskeysStartLoginResp{}, nil
}
