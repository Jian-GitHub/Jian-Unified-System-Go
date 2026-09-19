// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package thirdParty

import (
	"context"
	"jian-unified-system/apollo/apollo-api/internal/logic/response"
	"jian-unified-system/apollo/apollo-rpc/apollo"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CallbackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CallbackLogic {
	return &CallbackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CallbackLogic) Callback(req *types.CallbackReq) (resp *types.RegResp, err error) {
	code := req.Code
	if req.Error != "" {
		code = ""
	}
	r, err := l.svcCtx.ThirdParty.HandleCallback(l.ctx, &apollo.ThirdPartyContinueReq{Provider: req.Provider, State: req.State, Code: code})
	if err != nil {
		return nil, err
	}
	token, err := l.svcCtx.Sessions.Issue(l.ctx, r.UserId)
	if err != nil {
		return nil, err
	}
	return &types.RegResp{BaseResponse: response.OK(), Data: types.RegData{Token: token}}, nil
}
