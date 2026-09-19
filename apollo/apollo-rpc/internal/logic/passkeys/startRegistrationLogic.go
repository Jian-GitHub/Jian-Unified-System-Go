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

type StartRegistrationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewStartRegistrationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StartRegistrationLogic {
	return &StartRegistrationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *StartRegistrationLogic) StartRegistration(in *apollo.PasskeysStartRegistrationReq) (*apollo.PasskeysStartRegistrationResp, error) {
	if in == nil {
		return nil, response.Error(identity.ErrInvalid)
	}
	name := in.DisplayName
	if name == "" {
		name = in.UserName
	}
	if name == "" {
		name = "Apollo"
	}
	locale, language := account.NormalizeLocale(in.Locale), in.Language
	if language == "" {
		language = "en"
	}
	profile, err := account.NewMinimalProfile(in.UserId, locale, language)
	if err != nil {
		return nil, response.Error(err)
	}
	options, session, err := l.svcCtx.Passkey.StartRegistration(l.ctx, profile, name, in.Bind)
	if err != nil {
		return nil, response.Error(err)
	}
	return &apollo.PasskeysStartRegistrationResp{OptionsJson: options, SessionData: []byte(session)}, nil
}
