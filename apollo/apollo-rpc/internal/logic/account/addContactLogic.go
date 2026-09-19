package accountlogic

import (
	"context"
	"strconv"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
	"jian-unified-system/apollo/apollo-rpc/internal/logic/response"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddContactLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddContactLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddContactLogic {
	return &AddContactLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AddContactLogic) AddContact(in *apollo.AddContactReq) (*apollo.AddContactResp, error) {
	if in == nil {
		return nil, response.Error(identity.ErrInvalid)
	}
	contact, err := l.svcCtx.Account.AddContact(l.ctx, in.UserId, in.Value, in.Type, in.PhoneRegion)
	if err != nil {
		return nil, response.Error(err)
	}
	return &apollo.AddContactResp{Contact: &apollo.UserContact{Id: strconv.FormatInt(contact.ID(), 10), Value: contact.Value(), Type: contact.Type(), PhoneRegion: contact.PhoneRegion(), Primary: contact.Primary()}}, nil
}
