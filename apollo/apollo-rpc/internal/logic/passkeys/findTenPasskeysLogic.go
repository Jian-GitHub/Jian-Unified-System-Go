package passkeyslogic

import (
	"context"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
	"jian-unified-system/apollo/apollo-rpc/internal/logic/response"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type FindTenPasskeysLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFindTenPasskeysLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindTenPasskeysLogic {
	return &FindTenPasskeysLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FindTenPasskeysLogic) FindTenPasskeys(in *apollo.FindTenPasskeysReq) (*apollo.FindTenPasskeysResp, error) {
	if in == nil {
		return nil, response.Error(identity.ErrInvalid)
	}
	items, err := l.svcCtx.Passkey.List(l.ctx, in.UserId, in.Page)
	if err != nil {
		return nil, response.Error(err)
	}
	out := make([]*apollo.Passkey, 0, len(items))
	for _, p := range items {
		out = append(out, response.Passkey(p))
	}
	return &apollo.FindTenPasskeysResp{Passkeys: out}, nil
}
