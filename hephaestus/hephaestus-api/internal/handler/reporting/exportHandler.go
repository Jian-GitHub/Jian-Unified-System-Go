// Code scaffolded by goctl 1.9.2. Safe to edit: HTTP transport only.
package reporting

import (
	"jian-unified-system/hephaestus/hephaestus-api/internal/access"
	"jian-unified-system/hephaestus/hephaestus-api/internal/logic/reporting"
	"jian-unified-system/hephaestus/hephaestus-api/internal/svc"
	"jian-unified-system/hephaestus/hephaestus-api/internal/types"
	"net/http"
)

func ExportHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.QueryReq
		if e := access.Parse(r, &req); e != nil {
			access.Write(w, nil, e)
			return
		}
		if req.Limit == 0 {
			req.Limit = 50
		}
		if !r.URL.Query().Has("rate_bps") {
			req.RateBps = 2000
		}
		l := reporting.NewExportLogic(r.Context(), svcCtx)
		out, e := l.Export(&req)
		if e != nil {
			access.Write(w, nil, e)
			return
		}
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="income.csv"`)
		_, _ = w.Write([]byte(out.Content))
	}
}
