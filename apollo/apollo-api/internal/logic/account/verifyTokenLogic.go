// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package account

import (
	"context"
	"jian-unified-system/apollo/apollo-api/internal/application"
	"jian-unified-system/apollo/apollo-api/internal/logic/response"
	"jian-unified-system/apollo/apollo-rpc/apollo"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type VerifyTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVerifyTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyTokenLogic {
	return &VerifyTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VerifyTokenLogic) VerifyToken(req *types.Empty) (resp *types.OkResp, err error) {
	id, err := response.Subject(l.ctx)
	if err != nil {
		return nil, err
	}
	if _, err = l.svcCtx.Account.SessionVersion(l.ctx, &apollo.SessionVersionReq{UserId: id}); err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, application.ErrCredentials
		}
		return nil, err
	}
	return &types.OkResp{BaseResponse: response.OK(), Data: types.OkData{Ok: true}}, nil
}
