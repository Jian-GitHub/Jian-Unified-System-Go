// Code scaffolded by goctl 1.9.2. Safe to edit: HTTP transport only.
package records

import (
	"jian-unified-system/hephaestus/hephaestus-api/internal/access"
	"jian-unified-system/hephaestus/hephaestus-api/internal/logic/records"
	"jian-unified-system/hephaestus/hephaestus-api/internal/svc"
	"jian-unified-system/hephaestus/hephaestus-api/internal/types"
	"net/http"
)

func UpdateRecordHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateRecordReq
		if e := access.Parse(r, &req); e != nil {
			access.Write(w, nil, e)
			return
		}
		l := records.NewUpdateRecordLogic(r.Context(), svcCtx)
		out, e := l.UpdateRecord(&req)
		access.Write(w, out, e)
	}
}
