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

type RemoveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRemoveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveLogic {
	return &RemoveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RemoveLogic) Remove(req *types.RemoveThirdPartyReq) (resp *types.OkResp, err error) {
	id, err := response.Subject(l.ctx)
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.ThirdParty.Remove(l.ctx, &apollo.ThirdPartyRemoveReq{UserId: id, ThirdPartyId: req.ThirdPartyId})
	if err != nil {
		return nil, err
	}
	return &types.OkResp{BaseResponse: response.OK(), Data: types.OkData{Ok: true}}, nil
}
