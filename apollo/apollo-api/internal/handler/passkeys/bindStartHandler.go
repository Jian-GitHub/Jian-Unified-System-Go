// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package passkeys

import (
	"jian-unified-system/apollo/apollo-api/internal/application"
	"net/http"

	"jian-unified-system/apollo/apollo-api/internal/logic/passkeys"
	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func BindStartHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.BindStartReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, application.ErrInvalid)
			return
		}

		l := passkeys.NewBindStartLogic(r.Context(), svcCtx)
		resp, err := l.BindStart(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
