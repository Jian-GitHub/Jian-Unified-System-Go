// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package security

import (
	"jian-unified-system/apollo/apollo-api/internal/application"
	"net/http"

	"jian-unified-system/apollo/apollo-api/internal/logic/account/security"
	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetTenSubsystemTokensHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PageReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, application.ErrInvalid)
			return
		}

		l := security.NewGetTenSubsystemTokensLogic(r.Context(), svcCtx)
		resp, err := l.GetTenSubsystemTokens(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
