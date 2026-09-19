// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package sso

import (
	"net/http"

	"jian-unified-system/apollo/apollo-api/internal/logic/sso"
	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func SSORevokeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		r.Body = http.MaxBytesReader(w, r.Body, 16384)
		var req types.SSOTokenReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := sso.NewSSORevokeLogic(r.Context(), svcCtx)
		resp, err := l.SSORevoke(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
