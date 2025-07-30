package passkeyslogic

import (
	"context"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type StartRegistrationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewStartRegistrationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StartRegistrationLogic {
	return &StartRegistrationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 注册
func (l *StartRegistrationLogic) StartRegistration(in *apollo.PasskeysStartRegistrationReq) (*apollo.PasskeysStartRegistrationResp, error) {
	// todo: add your logic here and delete this line

	return &apollo.PasskeysStartRegistrationResp{}, nil
}
