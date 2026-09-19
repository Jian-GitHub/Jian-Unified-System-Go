package accountlogic

import (
	"context"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
	"jian-unified-system/apollo/apollo-rpc/internal/logic/response"
	"strconv"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserSecurityInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserSecurityInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserSecurityInfoLogic {
	return &UserSecurityInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UserSecurityInfoLogic) UserSecurityInfo(in *apollo.UserSecurityInfoReq) (*apollo.UserSecurityInfoResp, error) {
	if in == nil {
		return nil, response.Error(identity.ErrInvalid)
	}
	u, err := l.svcCtx.Passkey.Security(l.ctx, in.UserId)
	if err != nil {
		return nil, response.Error(err)
	}
	items := u.Contacts()
	contacts := make([]*apollo.UserContact, 0, len(items))
	for _, c := range items {
		contacts = append(contacts, &apollo.UserContact{Id: strconv.FormatInt(c.ID(), 10), Value: c.Value(), Type: c.Type(), PhoneRegion: c.PhoneRegion(), Primary: c.Primary()})
	}
	y, m, d := response.Date(u.PasswordUpdatedAt())
	return &apollo.UserSecurityInfoResp{Contacts: contacts, PasswordUpdatedDate: &apollo.PasswordUpdatedDate{Year: y, Month: m, Day: d}, AccountSecurityTokenNum: u.Tokens(), PasskeysNum: u.Passkeys(), ThirdPartyAccounts: &apollo.ThirdPartyAccounts{Github: u.Github(), Google: u.Google()}}, nil
}
