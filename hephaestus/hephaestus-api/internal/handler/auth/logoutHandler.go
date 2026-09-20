// Code scaffolded by goctl 1.9.2. Safe to edit: HTTP transport only.
package auth

import (
	"jian-unified-system/hephaestus/hephaestus-api/internal/access"
	"jian-unified-system/hephaestus/hephaestus-api/internal/logic/auth"
	"jian-unified-system/hephaestus/hephaestus-api/internal/svc"
	"jian-unified-system/hephaestus/hephaestus-api/internal/types"
	"net/http"
)

func LogoutHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.Empty
		if e := access.Parse(r, &req); e != nil {
			access.Write(w, nil, e)
			return
		}
		l := auth.NewLogoutLogic(r.Context(), svcCtx)
		out, e := l.Logout(&req)
		access.Write(w, out, e)
	}
}
