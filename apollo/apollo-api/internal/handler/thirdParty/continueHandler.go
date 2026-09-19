// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package thirdParty

import (
	"jian-unified-system/apollo/apollo-api/internal/application"
	"net/http"

	"jian-unified-system/apollo/apollo-api/internal/logic/thirdParty"
	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func ContinueHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ProviderReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, application.ErrInvalid)
			return
		}

		l := thirdParty.NewContinueLogic(r.Context(), svcCtx)
		resp, err := l.Continue(&req, svcCtx.Country(r))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			if err := beginBrowser(w, r, resp.Data.Url, svcCtx.Config.Auth.AccessSecret); err != nil {
				httpx.ErrorCtx(r.Context(), w, err)
				return
			}
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
