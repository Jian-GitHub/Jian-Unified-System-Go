// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package account

import (
	"context"
	"jian-unified-system/apollo/apollo-api/internal/logic/response"
	"jian-unified-system/apollo/apollo-rpc/apollo"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserSecurityInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserSecurityInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserSecurityInfoLogic {
	return &GetUserSecurityInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserSecurityInfoLogic) GetUserSecurityInfo(req *types.Empty) (resp *types.GetUserSecurityInfoResp, err error) {
	id, err := response.Subject(l.ctx)
	if err != nil {
		return nil, err
	}
	r, err := l.svcCtx.Account.UserSecurityInfo(l.ctx, &apollo.UserSecurityInfoReq{UserId: id})
	if err != nil {
		return nil, err
	}
	contacts := make([]types.UserContact, 0, len(r.Contacts))
	for _, c := range r.Contacts {
		contacts = append(contacts, types.UserContact{Id: c.Id, Value: c.Value, Type: c.Type, PhoneRegion: c.PhoneRegion, Primary: c.Primary})
	}
	return &types.GetUserSecurityInfoResp{BaseResponse: response.OK(), Data: types.SecurityData{Contacts: contacts, PasswordUpdatedDate: types.Birthday{Year: r.PasswordUpdatedDate.GetYear(), Month: r.PasswordUpdatedDate.GetMonth(), Day: r.PasswordUpdatedDate.GetDay()}, AccountSecurityTokenNum: r.AccountSecurityTokenNum, PasskeysNum: r.PasskeysNum, ThirdPartyAccounts: types.ThirdPartyAccounts{Github: r.ThirdPartyAccounts.GetGithub(), Google: r.ThirdPartyAccounts.GetGoogle()}}}, nil
}
