package oauth2

import (
	"encoding/json"
	"errors"
	"github.com/zeromicro/go-zero/core/errorx"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	ot "jian-unified-system/jus-core/types/oauth2"
)

// ParseThirdPartyUser 解析 第三方 响应数据, 返回 第三方用户
func ParseThirdPartyUser(endpoint oauth2.Endpoint, data []byte) (user ot.ThirdPartyUser, err error) {
	switch endpoint {
	case google.Endpoint:
		var userProfile *ot.GoogleUserProfile
		err = json.Unmarshal(data, &userProfile)
		if err != nil {
			return nil, errorx.Wrap(errors.New("parse third party user fail"), "ThirdParty Continue Err")
		}
		return &ot.GoogleAdapter{GoogleUserProfile: userProfile}, nil
	case github.Endpoint:
		var userInfo *ot.GitHubUserInfo
		err = json.Unmarshal(data, &userInfo)
		if err != nil {
			return nil, errorx.Wrap(errors.New("parse third party user fail"), "ThirdParty Continue Err")
		}
		return &ot.GitHubAdapter{GitHubUserInfo: userInfo}, nil
	default:
		return nil, errorx.Wrap(errors.New("unknown endpoint"), "ThirdParty Continue Err")
	}
}
