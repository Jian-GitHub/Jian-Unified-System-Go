// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package security

import (
	"context"
	"encoding/json"
	"jian-unified-system/apollo/apollo-api/internal/logic/response"
	"jian-unified-system/apollo/apollo-rpc/apollo"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateSubsystemTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGenerateSubsystemTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateSubsystemTokenLogic {
	return &GenerateSubsystemTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GenerateSubsystemTokenLogic) GenerateSubsystemToken(req *types.GenerateSubsystemTokenReq) (resp *types.SubsystemTokenResp, err error) {
	id, err := response.Subject(l.ctx)
	if err != nil {
		return nil, err
	}
	scopes, err := json.Marshal(req.Scope)
	if err != nil {
		return nil, err
	}
	r, err := l.svcCtx.Security.GenerateSubsystemToken(l.ctx, &apollo.GenerateSubsystemTokenReq{UserId: id, Name: req.Name, Scope: scopes})
	if err != nil {
		return nil, err
	}
	return &types.SubsystemTokenResp{BaseResponse: response.OK(), Data: types.SubsystemTokenData{Token: response.Grant(r.Token)}}, nil
}
