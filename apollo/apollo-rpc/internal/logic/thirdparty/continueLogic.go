package thirdpartylogic

import (
	"context"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ContinueLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewContinueLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ContinueLogic {
	return &ContinueLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 继续 - 登录或注册
func (l *ContinueLogic) Continue(in *apollo.ThirdPartyBindReq) (*apollo.ThirdPartyContinueResp, error) {
	// todo: add your logic here and delete this line

	return &apollo.ThirdPartyContinueResp{}, nil
}
