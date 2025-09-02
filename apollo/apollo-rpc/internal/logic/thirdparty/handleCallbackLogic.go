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
	"io"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"
	o2 "jian-unified-system/apollo/apollo-rpc/util/oauth2"
	ap "jian-unified-system/jus-core/data/mysql/apollo"
	o2util "jian-unified-system/jus-core/util/oauth2"
	redisUtil "jian-unified-system/jus-core/util/oauth2/redis"
	"strconv"

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
	// 1. Check Redis Cache
	// Parse Redis Data
	// redis data string -> RedisData
	fmt.Println(in.RedisDataJson)
	var redisData *redisUtil.RedisData
	err := json.Unmarshal([]byte(in.RedisDataJson), &redisData)
	if err != nil {
		return nil, errorx.Wrap(errors.New("redis state fail"), "ThirdParty Continue Err")
	}

	// 2. Check Token
	// Parse Token
	var token *oauth2.Token
	err = json.Unmarshal(in.Token, &token)
	if err != nil {
		return nil, errors.New("token no validated")
	}
	// 3. Check Provider
	config, ok := l.svcCtx.OauthProviders[in.Provider]
	if !ok {
		return nil, errorx.Wrap(errors.New("no provider"), "ThirdParty Continue Err")
	}

	// 4. OAuth2 --Token--> Response
	client := config.Client(context.Background(), token)
	resp, err := client.Get(config.UserInfoURL)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	_ = resp.Body.Close()

	thirdPartyUser, err := o2util.ParseThirdPartyUser(config.Endpoint, body)
	if err != nil {
		return nil, err
	}

	// set tp exist flag
	thirdPartyIsExist := false
	// find tp from database
	existedThirdParty, err := l.svcCtx.ThirdPartyModel.FindOneByThirdID(l.ctx, thirdPartyUser.GetID())
	switch {
	// tp exists
	case err == nil:
		thirdPartyIsExist = true
	// tp does not exist
	case errors.Is(err, sqlx.ErrNotFound):
		thirdPartyIsExist = false
	// database err
	default:
		return nil, err
	}

	fmt.Println("thirdPartyIsExist", thirdPartyIsExist)

	// 需要查 user 的情况
	// thirdParty exists: 		 p,		continue:	 q
	// thirdParty no existing:	¬p,		Bind:		¬q
	// 		1. p ∧ q
	// 		2. ¬p ∧ ¬q
	var user *ap.User
	var defaultUserId int64
	if thirdPartyIsExist {
		defaultUserId = existedThirdParty.UserId
	} else {
		defaultUserId, err = strconv.ParseInt(redisData.Id, 10, 64)
		if err != nil {
			return nil, errorx.Wrap(err, "ThirdParty Continue Err")
		}
	}

	switch {
	// (第三方存在 && Continue) 或 (第三方不存在 && Bind): 需要查 user
	case (thirdPartyIsExist && redisData.Flag == redisUtil.ContinueFlag) ||
		(!thirdPartyIsExist && redisData.Flag == redisUtil.BindFlag):
		// 查 user
		user, err = l.svcCtx.UserModel.FindOne(l.ctx, defaultUserId)
		switch {
		// user does not exist
		case errors.Is(err, sqlx.ErrNotFound):
			return nil, err
		// database err
		case err != nil:
			return nil, err
		}

		// 分情况讨论: 流程类型
		switch redisData.Flag {
		case redisUtil.ContinueFlag:
			fmt.Println("进入登录")
			// 处理 登录 - 返回 user defaultUserId
			// 更新第三方数据
			existedThirdParty.Name = thirdPartyUser.GetDisplayName()
			existedThirdParty.RawData = sql.NullString{
				String: string(body),
				Valid:  true,
			}
			_ = l.svcCtx.ThirdPartyModel.Update(l.ctx, existedThirdParty)
			return &apollo.ThirdPartyContinueResp{
				UserId: user.Id,
			}, nil
		case redisUtil.BindFlag:
			fmt.Println("进入绑定")
			// 处理 绑定 - 存新 third_party
			newThirdParty := &ap.ThirdParty{
				RawData: sql.NullString{
					String: string(body),
					Valid:  true,
				},
			}
			// 生成 third_party
			err = o2.ParseUserAndContacts(&thirdPartyUser, user, nil, newThirdParty)
			if err != nil {
				return nil, err
			}
			// 存入 third_party
			err = l.svcCtx.ThirdPartyModel.Update(l.ctx, newThirdParty)
			if err != nil {
				return nil, err
			}
			// 返回用户 id
			return &apollo.ThirdPartyContinueResp{
				UserId: user.Id,
			}, nil
		}
	// 第三方不存在 && Continue: 注册流程
	case !thirdPartyIsExist && redisData.Flag == redisUtil.ContinueFlag:
		fmt.Println("进入注册")
		// 处理注册
		// 创建 新 user, 新 contacts, 新 third_party
		newUser := &ap.User{
			Id: defaultUserId,
		}
		newContacts := make([]*ap.Contact, 0)
		newThirdParty := &ap.ThirdParty{
			RawData: sql.NullString{
				String: string(body),
				Valid:  true,
			},
		}

		err = o2.ParseUserAndContacts(&thirdPartyUser, newUser, &newContacts, newThirdParty)
		if err != nil {
			return nil, err
		}

		// 存入数据库
		// user
		_, err = l.svcCtx.UserModel.Insert(l.ctx, newUser)
		if err != nil {
			return nil, err
		}
		// contacts
		fmt.Println("检查 contacts")
		fmt.Println(newContacts == nil)
		fmt.Println(len(newContacts))
		if len(newContacts) > 0 {
			_, err = l.svcCtx.ContactModel.InsertBatch(l.ctx, newContacts)
			if err != nil {
				return nil, err
			}
		}
		// thirdParty
		_, err = l.svcCtx.ThirdPartyModel.Insert(l.ctx, newThirdParty)
		if err != nil {
			return nil, err
		}

		// 注册完成 - 返回用户 defaultUserId
		return &apollo.ThirdPartyContinueResp{
			UserId: newUser.Id,
		}, nil

	// 第三方存在 && Bind: Error
	case thirdPartyIsExist && redisData.Flag == redisUtil.BindFlag:
		//  第三方已经被绑定 - 异常
		return nil, errorx.Wrap(errors.New("already bound"), "ThirdParty Bind Err")
	}

	return nil, errorx.Wrap(errors.New("unknown flag"), "ThirdParty Continue Err")
}
