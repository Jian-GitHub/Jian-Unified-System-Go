package securitylogic

import (
	"context"
	"encoding/json"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
	"jian-unified-system/apollo/apollo-rpc/internal/logic/response"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateSubsystemTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGenerateSubsystemTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateSubsystemTokenLogic {
	return &GenerateSubsystemTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GenerateSubsystemTokenLogic) GenerateSubsystemToken(in *apollo.GenerateSubsystemTokenReq) (*apollo.GenerateSubsystemTokenResp, error) {
	if in == nil {
		return nil, response.Error(identity.ErrInvalid)
	}
	var scopes []int64
	if err := json.Unmarshal(in.Scope, &scopes); err != nil {
		return nil, response.Error(identity.ErrInvalid)
	}
	grant, err := l.svcCtx.Grant.Create(l.ctx, in.UserId, in.Name, scopes)
	if err != nil {
		return nil, response.Error(err)
	}
	return &apollo.GenerateSubsystemTokenResp{Token: response.Grant(grant)}, nil
}
