// Code scaffolded by goctl 1.9.2. Safe to edit: HTTP transport only.
package imports

import (
	"jian-unified-system/hephaestus/hephaestus-api/internal/access"
	"jian-unified-system/hephaestus/hephaestus-api/internal/logic/imports"
	"jian-unified-system/hephaestus/hephaestus-api/internal/svc"
	"jian-unified-system/hephaestus/hephaestus-api/internal/types"
	"net/http"
)

func ListRowsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.BatchRowsReq
		if e := access.Parse(r, &req); e != nil {
			access.Write(w, nil, e)
			return
		}
		l := imports.NewListRowsLogic(r.Context(), svcCtx)
		out, e := l.ListRows(&req)
		access.Write(w, out, e)
	}
}
