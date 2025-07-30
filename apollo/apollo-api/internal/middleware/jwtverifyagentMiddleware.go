package middleware

import (
	"fmt"
	"net/http"
)

type JWTVerifyAgentMiddleware struct {
}

func NewJWTVerifyAgentMiddleware() *JWTVerifyAgentMiddleware {
	return &JWTVerifyAgentMiddleware{}
}

func (m *JWTVerifyAgentMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO generate middleware implement function, delete after code implementation
		fmt.Println("进入中间件: JWTVerifyAgentMiddleware")
		// Passthrough to next handler if need
		next(w, r)
	}
}
