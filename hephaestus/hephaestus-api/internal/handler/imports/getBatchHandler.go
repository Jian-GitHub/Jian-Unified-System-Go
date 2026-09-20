// Code scaffolded by goctl 1.9.2. Safe to edit: HTTP transport only.
package imports

import (
	"jian-unified-system/hephaestus/hephaestus-api/internal/access"
	"jian-unified-system/hephaestus/hephaestus-api/internal/logic/imports"
	"jian-unified-system/hephaestus/hephaestus-api/internal/svc"
	"jian-unified-system/hephaestus/hephaestus-api/internal/types"
	"net/http"
)

func GetBatchHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.IDReq
		if e := access.Parse(r, &req); e != nil {
			access.Write(w, nil, e)
			return
		}
		l := imports.NewGetBatchLogic(r.Context(), svcCtx)
		out, e := l.GetBatch(&req)
		access.Write(w, out, e)
	}
}
