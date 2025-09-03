package job

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"jian-unified-system/jquantum/jquantum-api/internal/logic/job"
	"jian-unified-system/jquantum/jquantum-api/internal/svc"
	"jian-unified-system/jquantum/jquantum-api/internal/types"
)

func RetrieveResultHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RetrieveResultReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := job.NewRetrieveResultLogic(r.Context(), svcCtx)
		resp, err := l.RetrieveResult(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
