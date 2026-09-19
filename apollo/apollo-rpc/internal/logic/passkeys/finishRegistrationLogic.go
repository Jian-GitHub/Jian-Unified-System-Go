package passkeyslogic

import (
	"context"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/account"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
	"jian-unified-system/apollo/apollo-rpc/internal/logic/response"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type FinishRegistrationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFinishRegistrationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FinishRegistrationLogic {
	return &FinishRegistrationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FinishRegistrationLogic) FinishRegistration(in *apollo.PasskeysFinishRegistrationReq) (*apollo.PasskeysFinishRegistrationResp, error) {
	if in == nil {
		return nil, response.Error(identity.ErrInvalid)
	}
	in.Locate = account.NormalizeLocale(in.Locate)
	key, p, err := l.svcCtx.Passkey.FinishRegistration(l.ctx, string(in.SessionData), in.UserId, in.Type, in.CredentialJson, in.Locate, in.Language)
	if err != nil {
		return nil, response.Error(err)
	}
	y, m, d := response.Date(key.CreatedAt())
	return &apollo.PasskeysFinishRegistrationResp{PasskeysId: key.ID(), PasskeysName: key.Name(), Locale: p.Locale(), Language: p.Language(), Year: y, Month: m, Day: d, UserId: p.ID()}, nil
}
