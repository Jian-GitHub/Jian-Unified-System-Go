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

func RemoveHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RemoveThirdPartyReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, application.ErrInvalid)
			return
		}

		l := thirdParty.NewRemoveLogic(r.Context(), svcCtx)
		resp, err := l.Remove(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
