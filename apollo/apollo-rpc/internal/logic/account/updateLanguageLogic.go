package accountlogic

import (
	"context"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
	"jian-unified-system/apollo/apollo-rpc/internal/logic/response"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateLanguageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateLanguageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLanguageLogic {
	return &UpdateLanguageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateLanguageLogic) UpdateLanguage(in *apollo.UpdateLanguageReq) (*apollo.Empty, error) {
	if in == nil {
		return nil, response.Error(identity.ErrInvalid)
	}
	if err := l.svcCtx.Account.UpdateLanguage(l.ctx, in.UserId, in.Language); err != nil {
		return nil, response.Error(err)
	}
	return &apollo.Empty{}, nil
}
