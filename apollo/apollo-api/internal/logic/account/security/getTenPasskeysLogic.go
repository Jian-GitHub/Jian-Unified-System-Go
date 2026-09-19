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

type GetTenPasskeysLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTenPasskeysLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenPasskeysLogic {
	return &GetTenPasskeysLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTenPasskeysLogic) GetTenPasskeys(req *types.PageReq) (resp *types.PasskeysResp, err error) {
	id, err := response.Subject(l.ctx)
	if err != nil {
		return nil, err
	}
	r, err := l.svcCtx.Passkeys.FindTenPasskeys(l.ctx, &apollo.FindTenPasskeysReq{UserId: id, Page: req.Page})
	if err != nil {
		return nil, err
	}
	items := make([]types.Passkey, 0, len(r.Passkeys))
	for _, p := range r.Passkeys {
		items = append(items, types.Passkey{Id: p.Id, Name: p.Name, Date: types.Birthday{Year: p.Year, Month: p.Month, Day: p.Day}, IsEnabled: p.IsEnabled})
	}
	return &types.PasskeysResp{BaseResponse: response.OK(), Data: types.PasskeysData{Passkeys: items}}, nil
}
