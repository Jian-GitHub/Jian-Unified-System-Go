package jquantumlogic

import (
	"context"
	"jian-unified-system/apollo/apollo-rpc/apollo"

	"jian-unified-system/jquantum/jquantum-rpc/internal/svc"
	"jian-unified-system/jquantum/jquantum-rpc/jquantum"

	"github.com/zeromicro/go-zero/core/logx"
)

type ValidateTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewValidateTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateTokenLogic {
	return &ValidateTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ValidateTokenLogic) ValidateToken(in *jquantum.ValidateTokenReq) (*jquantum.ValidateTokenResp, error) {
	resp, err := l.svcCtx.ApolloAccount.ValidateToken(l.ctx, &apollo.ValidateTokenReq{
		TokenId: in.TokenId,
	})
	if err != nil {
		return nil, err
	}

	return &jquantum.ValidateTokenResp{
		Validated: resp.Validated,
	}, nil
}
