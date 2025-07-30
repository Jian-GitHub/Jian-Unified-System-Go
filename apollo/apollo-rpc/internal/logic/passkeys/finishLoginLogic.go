package passkeyslogic

import (
	"context"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type FinishLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFinishLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FinishLoginLogic {
	return &FinishLoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// FinishLogin 登陆第二步 - 完成 返回用户id
func (l *FinishLoginLogic) FinishLogin(in *apollo.PasskeysFinishLoginReq) (*apollo.PasskeysFinishLoginResp, error) {
	// todo: add your logic here and delete this line

	return &apollo.PasskeysFinishLoginResp{}, nil
}
