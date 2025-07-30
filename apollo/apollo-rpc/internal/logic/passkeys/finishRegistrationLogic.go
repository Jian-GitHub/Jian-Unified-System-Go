package passkeyslogic

import (
	"context"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type FinishRegistrationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFinishRegistrationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FinishRegistrationLogic {
	return &FinishRegistrationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// FinishRegistration 注册第二步 - 完成
func (l *FinishRegistrationLogic) FinishRegistration(in *apollo.PasskeysFinishRegistrationReq) (*apollo.Empty, error) {
	// todo: add your logic here and delete this line

	return &apollo.Empty{}, nil
}
