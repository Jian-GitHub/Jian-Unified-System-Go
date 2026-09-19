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

type RemovePasskeyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRemovePasskeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemovePasskeyLogic {
	return &RemovePasskeyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RemovePasskeyLogic) RemovePasskey(req *types.RemoveCredentialReq) (resp *types.OkResp, err error) {
	id, err := response.Subject(l.ctx)
	if err != nil {
		return nil, err
	}
	r, err := l.svcCtx.Passkeys.RemovePasskey(l.ctx, &apollo.RemovePasskeyReq{UserId: id, PasskeyId: req.Id})
	if err != nil {
		return nil, err
	}
	return &types.OkResp{BaseResponse: response.OK(), Data: types.OkData{Ok: r.Success}}, nil
}
