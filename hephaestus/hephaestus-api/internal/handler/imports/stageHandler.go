// Code scaffolded by goctl 1.9.2. Safe to edit: HTTP transport only.
package imports

import (
	"jian-unified-system/hephaestus/hephaestus-api/internal/access"
	"jian-unified-system/hephaestus/hephaestus-api/internal/logic/imports"
	"jian-unified-system/hephaestus/hephaestus-api/internal/svc"
	"jian-unified-system/hephaestus/hephaestus-api/internal/types"
	"net/http"
)

func StageHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		source, cleanup, e := svcCtx.Access.Upload(r)
		defer cleanup()
		if e != nil {
			access.Write(w, nil, e)
			return
		}
		req := types.StageReq{SourceId: source}
		l := imports.NewStageLogic(r.Context(), svcCtx)
		out, e := l.Stage(&req)
		access.Write(w, out, e)
	}
}
