package thirdParty

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"jian-unified-system/apollo/apollo-api/internal/logic/thirdParty"
	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"
)

func GetInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.Empty
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := thirdParty.NewGetInfoLogic(r.Context(), svcCtx)
		resp, err := l.GetInfo()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
