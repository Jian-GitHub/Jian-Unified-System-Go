package apolloUtil

import (
	"errors"
	"github.com/zeromicro/go-zero/core/errorx"
	"jian-unified-system/apollo/apollo-api/internal/svc"
)

func RedirectToOAuth2(svcCtx *svc.ServiceContext, provider string, state string) (string, error) {
	cfg, ok := svcCtx.OauthProviders[provider]
	if !ok {
		return "", errorx.Wrap(errors.New("no provider"), "ThirdParty Continue Err")
	}
	url := cfg.AuthCodeURL(state)
	return url, nil
}
