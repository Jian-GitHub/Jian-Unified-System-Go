// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package security

import (
	"context"
	"jian-unified-system/apollo/apollo-api/internal/logic/response"
	"jian-unified-system/apollo/apollo-rpc/apollo"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveSubsystemTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRemoveSubsystemTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveSubsystemTokenLogic {
	return &RemoveSubsystemTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RemoveSubsystemTokenLogic) RemoveSubsystemToken(req *types.RemoveCredentialReq) (resp *types.OkResp, err error) {
	id, err := response.Subject(l.ctx)
	if err != nil {
		return nil, err
	}
	key, err := response.ID(req.Id)
	if err != nil {
		return nil, err
	}
	r, err := l.svcCtx.Security.RemoveSubsystemToken(l.ctx, &apollo.RemoveSubsystemTokenReq{UserId: id, TokenId: key})
	if err != nil {
		return nil, err
	}
	return &types.OkResp{BaseResponse: response.OK(), Data: types.OkData{Ok: r.Validated}}, nil
}
