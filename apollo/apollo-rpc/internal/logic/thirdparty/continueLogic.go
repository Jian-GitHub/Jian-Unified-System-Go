package thirdpartylogic

import (
	"context"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
	"jian-unified-system/apollo/apollo-rpc/internal/logic/response"

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

func (l *ContinueLogic) Continue(in *apollo.ThirdPartyContinueReq) (*apollo.ThirdPartyContinueResp, error) {
	if in == nil {
		return nil, response.Error(identity.ErrInvalid)
	}
	id, err := l.svcCtx.OAuth.CompleteAuthorization(l.ctx, in.Provider, in.State, in.Code, "login", 0)
	if err != nil {
		return nil, response.Error(err)
	}
	return &apollo.ThirdPartyContinueResp{UserId: id}, nil
}
