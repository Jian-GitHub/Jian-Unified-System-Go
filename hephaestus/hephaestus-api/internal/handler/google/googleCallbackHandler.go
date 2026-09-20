// Code scaffolded by goctl. Safe to edit: HTTP transport only.
package google

import (
	"jian-unified-system/hephaestus/hephaestus-api/internal/access"
	logic "jian-unified-system/hephaestus/hephaestus-api/internal/logic/google"
	"jian-unified-system/hephaestus/hephaestus-api/internal/svc"
	"jian-unified-system/hephaestus/hephaestus-api/internal/types"
	"net/http"
)

func GoogleCallbackHandler(s *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GoogleCallbackReq
		if e := access.Parse(r, &req); e != nil {
			access.Write(w, nil, e)
			return
		}
		out, e := logic.NewGoogleCallbackLogic(r.Context(), s).GoogleCallback(&req)
		_ = out
		if e != nil {
			access.Write(w, nil, e)
		}
	}
}
