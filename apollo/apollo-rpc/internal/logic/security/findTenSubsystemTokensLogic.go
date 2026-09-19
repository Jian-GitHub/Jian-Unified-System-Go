package securitylogic

import (
	"context"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
	"jian-unified-system/apollo/apollo-rpc/internal/logic/response"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type FindTenSubsystemTokensLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFindTenSubsystemTokensLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindTenSubsystemTokensLogic {
	return &FindTenSubsystemTokensLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FindTenSubsystemTokensLogic) FindTenSubsystemTokens(in *apollo.FindTenSubsystemTokensReq) (*apollo.FindTenSubsystemTokensResp, error) {
	if in == nil {
		return nil, response.Error(identity.ErrInvalid)
	}
	items, err := l.svcCtx.Grant.List(l.ctx, in.UserId, in.Page)
	if err != nil {
		return nil, response.Error(err)
	}
	out := make([]*apollo.SubsystemToken, 0, len(items))
	for _, g := range items {
		out = append(out, response.Grant(g))
	}
	return &apollo.FindTenSubsystemTokensResp{Tokens: out}, nil
}
