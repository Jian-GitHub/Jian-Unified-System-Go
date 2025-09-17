package account

import (
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-zero/core/jsonx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"io"
	"jian-unified-system/apollo/apollo-api/internal/logic/account"
	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"
	"net/http"
	"net/url"
)

func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		type TurnstileResponse struct {
			Success     bool     `json:"success"`
			ChallengeTs string   `json:"challenge_ts"`
			Hostname    string   `json:"hostname"`
			Action      string   `json:"action"` // 新增字段
			ErrorCodes  []string `json:"error-codes"`
		}

		if req.CloudflareToken == "" {
			return
		}

		// 设置 CORS 头
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// 或开发阶段临时允许所有（不推荐生产环境）
		// w.Header().Set("Access-Control-Allow-Origin", "*")

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// 处理预检请求 (OPTIONS)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		fmt.Println(req.CloudflareToken)

		form := url.Values{}
		form.Add("secret", "0x4AAAAAAANVWcsutYm3MqUvm50PJPaBi3s")
		form.Add("response", req.CloudflareToken)
		respCloudflare, err := http.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify", form)
		if err != nil {
			fmt.Println(err.Error())
		}
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(respCloudflare.Body)

		var result TurnstileResponse
		if err := json.NewDecoder(respCloudflare.Body).Decode(&result); err != nil {
			fmt.Println(err.Error())
		}

		marshalToString, _ := jsonx.MarshalToString(result)
		fmt.Println(marshalToString)

		// 0x4AAAAAAANVWcsutYm3MqUvm50PJPaBi3s

		l := account.NewLoginLogic(r.Context(), svcCtx)

		resp, err := l.Login(&req /*, r*/)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
