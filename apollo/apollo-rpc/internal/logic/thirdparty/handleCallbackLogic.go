package thirdpartylogic

import (
	"context"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
	"jian-unified-system/apollo/apollo-rpc/internal/logic/response"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleCallbackLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHandleCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleCallbackLogic {
	return &HandleCallbackLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *HandleCallbackLogic) HandleCallback(in *apollo.ThirdPartyContinueReq) (*apollo.ThirdPartyContinueResp, error) {
	if in == nil {
		return nil, response.Error(identity.ErrInvalid)
	}
	id, err := l.svcCtx.OAuth.CompleteAuthorization(l.ctx, in.Provider, in.State, in.Code, "", 0)
	if err != nil {
		return nil, response.Error(err)
	}
	return &apollo.ThirdPartyContinueResp{UserId: id}, nil
}
