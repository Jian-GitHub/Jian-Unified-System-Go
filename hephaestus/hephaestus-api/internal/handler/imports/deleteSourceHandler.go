// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package imports

import (
	"net/http"

	"jian-unified-system/hephaestus/hephaestus-api/internal/access"
	"jian-unified-system/hephaestus/hephaestus-api/internal/logic/imports"
	"jian-unified-system/hephaestus/hephaestus-api/internal/svc"
	"jian-unified-system/hephaestus/hephaestus-api/internal/types"
)

func DeleteSourceHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.IDReq
		if err := access.Parse(r, &req); err != nil {
			access.Write(w, nil, err)
			return
		}

		l := imports.NewDeleteSourceLogic(r.Context(), svcCtx)
		resp, err := l.DeleteSource(&req)
		access.Write(w, resp, err)
	}
}
