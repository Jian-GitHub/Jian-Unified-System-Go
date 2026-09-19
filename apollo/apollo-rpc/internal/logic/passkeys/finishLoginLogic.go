package passkeyslogic

import (
	"context"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
	"jian-unified-system/apollo/apollo-rpc/internal/logic/response"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type FinishLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFinishLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FinishLoginLogic {
	return &FinishLoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FinishLoginLogic) FinishLogin(in *apollo.PasskeysFinishLoginReq) (*apollo.PasskeysFinishLoginResp, error) {
	if in == nil {
		return nil, response.Error(identity.ErrInvalid)
	}
	p, err := l.svcCtx.Passkey.FinishLogin(l.ctx, in.SessionDataJson, []byte(in.CredentialJson))
	if err != nil {
		return nil, response.Error(err)
	}
	return &apollo.PasskeysFinishLoginResp{UserId: p.ID(), GivenName: p.GivenName(), MiddleName: p.MiddleName(), FamilyName: p.FamilyName(), Avatar: p.Avatar(), Locale: p.Locale(), Language: p.Language(), BirthdayYear: p.BirthdayYear(), BirthdayMonth: p.BirthdayMonth(), BirthdayDay: p.BirthdayDay()}, nil
}
