package sso

import (
	"jian-unified-system/apollo/apollo-api/internal/types"
	"jian-unified-system/apollo/apollo-rpc/apollo"
)

func sessionResponse(v *apollo.SSOSessionResp, e error) (*types.SSOSessionResp, error) {
	if e != nil {
		return nil, e
	}
	return &types.SSOSessionResp{Token: v.Token, Subject: v.Subject, DisplayName: v.DisplayName, CsrfToken: v.CsrfToken, ExpiresAt: v.ExpiresAt, Scope: v.Scope}, nil
}
