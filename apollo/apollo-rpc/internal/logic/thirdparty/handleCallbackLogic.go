package thirdpartylogic

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"io"
	"jian-unified-system/apollo/apollo-rpc/internal/model"
	githubUtil "jian-unified-system/apollo/apollo-rpc/util/oauth2/github"
	googleUtil "jian-unified-system/apollo/apollo-rpc/util/oauth2/google"
	ot "jian-unified-system/jus-core/types/oauth2"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleCallbackLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHandleCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleCallbackLogic {
	return &HandleCallbackLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// HandleCallback 处理第三方回调数据
func (l *HandleCallbackLogic) HandleCallback(in *apollo.ThirdPartyContinueReq) (*apollo.ThirdPartyContinueResp, error) {
	// todo: add your logic here and delete this line
	// 1. Parse Token
	var token *oauth2.Token
	err := json.Unmarshal(in.Token, &token)
	if err != nil {
		return nil, errors.New("token no validated")
	}

	// 2. Check Provider
	config, ok := l.svcCtx.OauthProviders[in.Provider]
	if !ok {
		return nil, errorx.Wrap(errors.New("no provider"), "ThirdParty Continue Err")
	}
	// 3. OAuth2 --Token--> Response
	client := config.Client(context.Background(), token)
	resp, err := client.Get(config.UserInfoURL)
	if err != nil {
		return nil, err
	}
	//defer func(Body io.ReadCloser) {
	//	_ = Body.Close()
	//}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	_ = resp.Body.Close()

	user := &model.User{}
	var contacts []*model.Contact
	thirdParty := &model.ThirdParty{
		RawData: sql.NullString{
			String: string(body),
			Valid:  true,
		},
	}
	// 4. Response -> User Info
	switch config.Endpoint {
	case github.Endpoint:
		var userInfo *ot.GitHubUserInfo
		err = json.Unmarshal(body, &userInfo)
		if err != nil {
			return nil, err
		}
		//var userProfile *ot.GitHubUserProfile
		//err = json.Unmarshal(body, &userProfile)
		//if err != nil {
		//	return nil, err
		//}

		thirdParty.ThirdId = userInfo.ID

		thirdParty.Name = userInfo.Login
		if in.Id != 0 {
			user, contacts, err = githubUtil.ParseGitHubUserInfo(userInfo)
			if err != nil {
				return nil, err
			}
		}

	case google.Endpoint:
		var userProfile *ot.GoogleUserProfile
		//err = json.NewDecoder(resp.Body).Decode(&userProfile)
		err = json.Unmarshal(body, &userProfile)
		err = userProfile.SetGoogleID()
		if err != nil {
			return nil, err
		}

		thirdParty.ThirdId = userProfile.ID

		// Update Google Account Name
		for _, name := range userProfile.Names {
			if name.Metadata.Primary {
				thirdParty.Name = name.DisplayName
			}
		}
		if in.Id != 0 {
			user, contacts, err = googleUtil.ParseUserProfile(userProfile)
			if err != nil {
				return nil, err
			}
		}

		//fmt.Println(firstName, lastName)
	}

	// Check ID
	switch in.Id {
	// Login
	case 0:
		// Find User
		tp, err := l.svcCtx.ThirdPartyModel.FindOneByThirdID(l.ctx, thirdParty.ThirdId)
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errorx.Wrap(errors.New("No account"), "ThirdParty Continue Err")
		} else if err != nil {
			return nil, errorx.Wrap(errors.New("Databse err"), "ThirdParty Continue Err")
		}
		user.Id = tp.UserId

		// Update Third-Party Account Info
		tp.Name = thirdParty.Name
		tp.RawData = thirdParty.RawData
		_ = l.svcCtx.ThirdPartyModel.Update(l.ctx, tp)

	// Reg / Bind
	default:
		if user == nil {
			return nil, errorx.Wrap(errors.New("callback data no validated: user"), "ThirdParty Continue Err")
		}
		if thirdParty.Id == 0 {
			return nil, errorx.Wrap(errors.New("callback data no validated: thirdParty"), "ThirdParty Continue Err")
		}
		//if contacts == nil || len(contacts) == 0 {
		//	return nil, errorx.Wrap(errors.New("callback data no validated: contacts"), "ThirdParty Continue Err")
		//}

		user.Id = in.Id
		thirdParty.UserId = user.Id
		// 插入 新用户
		_, err = l.svcCtx.UserModel.Insert(l.ctx, user)
		if err != nil {
			return nil, errorx.Wrap(err, "ThirdParty Continue Err - insert new user fail")
		}

		// 插入 第三方账户
		_, err = l.svcCtx.ThirdPartyModel.Insert(l.ctx, thirdParty)
		if err != nil {
			return nil, errorx.Wrap(err, "ThirdParty Continue Err - insert new ThirdParty fail")
		}

		// 插入 经过验证的 电子邮件 -> contact 表
		if contacts != nil && len(contacts) > 0 {
			_, err = l.svcCtx.ContactModel.InsertBatch(l.ctx, contacts)
			if err != nil {
				return nil, errorx.Wrap(err, "ThirdParty Continue Err - insert new Contact fail")
			}
		}

	}

	return &apollo.ThirdPartyContinueResp{
		UserId: user.Id,
	}, nil
}
