package apolloUtil

import (
	"errors"
	"github.com/zeromicro/go-zero/core/errorx"
	"jian-unified-system/apollo/apollo-api/internal/svc"
	"net/http"
)

func RedirectToOAuth2(svcCtx *svc.ServiceContext, provider string, state string, w http.ResponseWriter, r *http.Request) error {
	cfg, ok := svcCtx.OauthProviders[provider]
	if !ok {
		return errorx.Wrap(errors.New("no provider"), "ThirdParty Continue Err")
	}
	url := cfg.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusFound)
	return nil
}
