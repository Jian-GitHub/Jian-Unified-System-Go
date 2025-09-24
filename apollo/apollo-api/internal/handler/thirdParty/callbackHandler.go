package thirdParty

import (
	"fmt"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"jian-unified-system/apollo/apollo-api/internal/logic/thirdParty"
	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"
)

func CallbackHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CallbackReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := thirdParty.NewCallbackLogic(r.Context(), svcCtx)
		resp, err := l.Callback(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			// 登录成功，重定向到前端
			frontendRedirectUrl := fmt.Sprintf("http://localhost:20551/login?token=%s", resp.CallbackRespData.Token)
			fmt.Println(frontendRedirectUrl)
			http.Redirect(w, r, frontendRedirectUrl, http.StatusFound)

			//httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
