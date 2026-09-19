package securitylogic

import (
	"context"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
	"jian-unified-system/apollo/apollo-rpc/internal/logic/response"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ValidateSubsystemTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewValidateSubsystemTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateSubsystemTokenLogic {
	return &ValidateSubsystemTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ValidateSubsystemTokenLogic) ValidateSubsystemToken(in *apollo.ValidateSubsystemTokenReq) (*apollo.ValidateSubsystemTokenResp, error) {
	if in == nil {
		return nil, response.Error(identity.ErrInvalid)
	}
	valid, err := l.svcCtx.Grant.Validate(l.ctx, in.UserId, in.TokenId)
	if err != nil {
		return nil, response.Error(err)
	}
	return &apollo.ValidateSubsystemTokenResp{Validated: valid}, nil
}
