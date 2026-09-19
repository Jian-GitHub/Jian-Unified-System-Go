// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package account

import (
	"context"

	"jian-unified-system/apollo/apollo-api/internal/logic/response"
	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"
	"jian-unified-system/apollo/apollo-rpc/apollo"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveNotificationEmailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRemoveNotificationEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveNotificationEmailLogic {
	return &RemoveNotificationEmailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RemoveNotificationEmailLogic) RemoveNotificationEmail(req *types.Empty) (resp *types.OkResp, err error) {
	id, err := response.Subject(l.ctx)
	if err != nil {
		return nil, err
	}
	if _, err = l.svcCtx.Account.RemoveNotificationEmail(l.ctx, &apollo.RemoveNotificationEmailReq{UserId: id}); err != nil {
		return nil, err
	}
	return &types.OkResp{BaseResponse: response.OK(), Data: types.OkData{Ok: true}}, nil
}
