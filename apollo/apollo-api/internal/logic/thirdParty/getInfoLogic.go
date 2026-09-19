// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package thirdParty

import (
	"context"
	"jian-unified-system/apollo/apollo-api/internal/logic/response"
	"jian-unified-system/apollo/apollo-rpc/apollo"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetInfoLogic {
	return &GetInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetInfoLogic) GetInfo(req *types.Empty) (resp *types.ThirdPartyResp, err error) {
	id, err := response.Subject(l.ctx)
	if err != nil {
		return nil, err
	}
	r, err := l.svcCtx.ThirdParty.GetInfo(l.ctx, &apollo.ThirdPartyGetInfoReq{UserId: id})
	if err != nil {
		return nil, err
	}
	items := make([]types.ThirdPartyAccount, 0, len(r.Accounts))
	for _, a := range r.Accounts {
		items = append(items, types.ThirdPartyAccount{Id: a.Id, Provider: a.Provider, Content: a.Content})
	}
	return &types.ThirdPartyResp{BaseResponse: response.OK(), Data: types.ThirdPartyData{Accounts: items}}, nil
}
