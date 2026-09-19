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

type GetTenSubsystemTokensLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTenSubsystemTokensLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenSubsystemTokensLogic {
	return &GetTenSubsystemTokensLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTenSubsystemTokensLogic) GetTenSubsystemTokens(req *types.PageReq) (resp *types.SubsystemTokensResp, err error) {
	id, err := response.Subject(l.ctx)
	if err != nil {
		return nil, err
	}
	r, err := l.svcCtx.Security.FindTenSubsystemTokens(l.ctx, &apollo.FindTenSubsystemTokensReq{UserId: id, Page: req.Page})
	if err != nil {
		return nil, err
	}
	items := make([]types.SubsystemToken, 0, len(r.Tokens))
	for _, g := range r.Tokens {
		items = append(items, response.Grant(g))
	}
	return &types.SubsystemTokensResp{BaseResponse: response.OK(), Data: types.SubsystemTokensData{Tokens: items}}, nil
}
