package thirdParty

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"jian-unified-system/apollo/apollo-api/internal/logic/thirdParty"
	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"
)

func RemoveHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RemoveReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
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
