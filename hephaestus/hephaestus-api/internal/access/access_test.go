package access

import (
	"context"
	"jian-unified-system/hephaestus/hephaestus-api/internal/types"
	"jian-unified-system/hephaestus/hephaestus-rpc/income"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGeneratedFormFieldsReachRPC(t *testing.T) {
	ctx := context.WithValue(context.Background(), key{}, &Meta{Owner: 17, Token: "trusted", Version: 3})
	req, e := Request[income.SummaryRequest](ctx, &types.QueryReq{RateBps: 2000, GroupBy: "week", Limit: 50, From: "2026-08-01"})
	if e != nil || req.OwnerId != 17 || req.Input.RateBps != 2000 || req.Input.GroupBy != "week" || req.Input.From != "2026-08-01" {
		t.Fatalf("mapping lost generated fields: %v %v", req, e)
	}
}
func TestRateLimitAndRejectedOrigin(t *testing.T) {
	m := Manager{Origins: []string{"http://localhost:15173"}}
	for i := 0; i < 10; i++ {
		if !m.allow("local") {
			t.Fatal("early rate rejection")
		}
	}
	if m.allow("local") {
		t.Fatal("rate limit not enforced")
	}
	r := httptest.NewRequest("POST", "/api/v1/session", nil)
	r.Header.Set("Origin", "https://evil.example")
	w := httptest.NewRecorder()
	m.Middleware(func(http.ResponseWriter, *http.Request) { t.Fatal("untrusted origin reached handler") })(w, r)
	if w.Code != 403 {
		t.Fatal(w.Code)
	}
}
