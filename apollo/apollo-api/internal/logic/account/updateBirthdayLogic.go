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

type UpdateBirthdayLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateBirthdayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateBirthdayLogic {
	return &UpdateBirthdayLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateBirthdayLogic) UpdateBirthday(req *types.UpdateBirthdayReq) (resp *types.OkResp, err error) {
	id, err := response.Subject(l.ctx)
	if err != nil {
		return nil, err
	}
	if _, err = l.svcCtx.Account.UpdateBirthday(l.ctx, &apollo.UpdateBirthdayReq{UserId: id, Year: req.Year, Month: req.Month, Day: req.Day}); err != nil {
		return nil, err
	}
	return &types.OkResp{BaseResponse: response.OK(), Data: types.OkData{Ok: true}}, nil
}
