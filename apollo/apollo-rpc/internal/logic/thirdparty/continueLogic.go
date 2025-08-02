package thirdpartylogic

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"io"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"
	googleUtil "jian-unified-system/apollo/apollo-rpc/util/oauth2/google"
	ot "jian-unified-system/jus-core/types/oauth2"
)

type ContinueLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewContinueLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ContinueLogic {
	return &ContinueLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 继续 - 登录或注册
func (l *ContinueLogic) Continue(in *apollo.ThirdPartyContinueReq) (*apollo.ThirdPartyContinueResp, error) {
	// todo: add your logic here and delete this line
	// 1. Parse Token
	var token *oauth2.Token
	err := json.Unmarshal(in.Token, &token)
	if err != nil {
		return nil, errors.New("token no validated")
	}

	// 2. OAuth2 --Token--> Response
	config := l.svcCtx.OauthProviders[in.Provider]
	client := config.Client(context.Background(), token)
	resp, err := client.Get(config.UserInfoURL)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// 3. Response -> User Info
	switch config.Endpoint {
	case github.Endpoint:
		var userInfo *ot.GitHubUserInfo
		err = json.NewDecoder(resp.Body).Decode(&userInfo)

	case google.Endpoint:
		var userProfile *ot.GoogleUserProfile
		err = json.NewDecoder(resp.Body).Decode(&userProfile)

		//fmt.Println(firstName, lastName)
		user, contacts, err := googleUtil.ParseUserProfile(userProfile)
		if err != nil {
			return nil, err
		}
		// 插入 新用户
		_, err = l.svcCtx.UserModel.Insert(l.ctx, user)
		if err != nil {
			return nil, err
		}
		// 插入 经过验证的 电子邮件 -> contact 表
		_, err = l.svcCtx.ContactModel.InsertBatch(l.ctx, contacts)
		if err != nil {
			return nil, err
		}

	}
	return &apollo.ThirdPartyContinueResp{
		UserId: 1,
	}, nil
}
