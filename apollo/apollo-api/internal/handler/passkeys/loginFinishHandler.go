package passkeys

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"jian-unified-system/apollo/apollo-api/internal/logic/passkeys"
	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"
)

func LoginFinishHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginFinishReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := passkeys.NewLoginFinishLogic(r.Context(), svcCtx)
		resp, err := l.LoginFinish(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
