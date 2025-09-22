package passkeys

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"jian-unified-system/apollo/apollo-api/internal/logic/passkeys"
	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"
)

func BindFinishHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.BindFinishReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := passkeys.NewBindFinishLogic(r.Context(), svcCtx)
		resp, err := l.BindFinish(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
