// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package thirdParty

import (
	"jian-unified-system/apollo/apollo-api/internal/application"
	"net/http"
	"net/url"

	"jian-unified-system/apollo/apollo-api/internal/logic/thirdParty"
	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func CallbackHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CallbackReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, application.ErrInvalid)
			return
		}

		if err := consumeBrowser(w, r, req.State, svcCtx.Config.Auth.AccessSecret); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := thirdParty.NewCallbackLogic(r.Context(), svcCtx)
		resp, err := l.Callback(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			target, err := url.Parse(svcCtx.Config.FrontendURL)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, err)
				return
			}
			query := target.Query()
			query.Set("token", resp.Data.Token)
			target.RawQuery = query.Encode()
			w.Header().Set("Cache-Control", "no-store")
			w.Header().Set("Referrer-Policy", "no-referrer")
			http.Redirect(w, r, target.String(), http.StatusFound)
		}
	}
}
