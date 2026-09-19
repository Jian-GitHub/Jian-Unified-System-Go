// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package account

import (
	"context"

	"jian-unified-system/apollo/apollo-api/internal/application"
	"jian-unified-system/apollo/apollo-api/internal/logic/response"
	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"
	"jian-unified-system/apollo/apollo-rpc/apollo"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddContactLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddContactLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddContactLogic {
	return &AddContactLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddContactLogic) AddContact(req *types.AddContactReq) (resp *types.AddContactResp, err error) {
	id, err := response.Subject(l.ctx)
	if err != nil {
		return nil, err
	}
	r, err := l.svcCtx.Account.AddContact(l.ctx, &apollo.AddContactReq{UserId: id, Value: req.Value, Type: req.Type, PhoneRegion: req.PhoneRegion})
	if err != nil {
		return nil, err
	}
	if r.Contact == nil {
		return nil, application.ErrInvalid
	}
	contact := r.Contact
	return &types.AddContactResp{BaseResponse: response.OK(), Data: types.ContactData{Contact: types.UserContact{Id: contact.Id, Value: contact.Value, Type: contact.Type, PhoneRegion: contact.PhoneRegion, Primary: contact.Primary}}}, nil
}
