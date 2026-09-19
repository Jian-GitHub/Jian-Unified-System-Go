package thirdpartylogic

import (
	"context"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
	"jian-unified-system/apollo/apollo-rpc/internal/logic/response"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetInfoLogic {
	return &GetInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetInfoLogic) GetInfo(in *apollo.ThirdPartyGetInfoReq) (*apollo.ThirdPartyGetInfoResp, error) {
	if in == nil {
		return nil, response.Error(identity.ErrInvalid)
	}
	items, err := l.svcCtx.OAuth.ExternalAccounts(l.ctx, in.UserId)
	if err != nil {
		return nil, response.Error(err)
	}
	out := make([]*apollo.ThirdPartyAccountInfo, 0, len(items))
	for _, a := range items {
		out = append(out, &apollo.ThirdPartyAccountInfo{Id: a.ID(), Provider: a.Provider(), Content: a.Content()})
	}
	return &apollo.ThirdPartyGetInfoResp{Accounts: out}, nil
}
