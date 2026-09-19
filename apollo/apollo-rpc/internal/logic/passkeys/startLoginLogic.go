package passkeyslogic

import (
	"context"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
	"jian-unified-system/apollo/apollo-rpc/internal/logic/response"

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

func (l *StartLoginLogic) StartLogin(in *apollo.Empty) (*apollo.PasskeysStartLoginResp, error) {
	if in == nil {
		return nil, response.Error(identity.ErrInvalid)
	}
	options, session, err := l.svcCtx.Passkey.StartLogin(l.ctx)
	if err != nil {
		return nil, response.Error(err)
	}
	return &apollo.PasskeysStartLoginResp{OptionsJson: options, SessionData: []byte(session)}, nil
}
