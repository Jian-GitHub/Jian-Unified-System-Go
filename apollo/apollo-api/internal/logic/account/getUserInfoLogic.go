package account

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/jsonx"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	ap "jian-unified-system/jus-core/data/mysql/apollo"
	"strconv"

	"jian-unified-system/apollo/apollo-api/internal/svc"
	"jian-unified-system/apollo/apollo-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserInfoLogic) GetUserInfo( /*req *types.GetUserInfoReq*/ ) (resp *types.GetUserInfoResp, err error) {
	id, err := l.ctx.Value("id").(json.Number).Int64()
	if err != nil {
		return &types.GetUserInfoResp{
			BaseResponse: types.BaseResponse{
				Code:    -1,
				Message: "Id err",
			},
		}, errorx.Wrap(errors.New("id"), "caller err")
	}

	userResp, err := l.svcCtx.ApolloAccount.UserInfo(l.ctx, &apollo.UserInfoReq{
		UserId: id,
	})
	if err != nil {
		return nil, err
	}
	var user ap.User

	err = jsonx.Unmarshal(userResp.UserBytes, &user)
	if err != nil {
		return &types.GetUserInfoResp{
			BaseResponse: types.BaseResponse{
				Code:    -2,
				Message: "user json unmarshal err",
			},
		}, errorx.Wrap(err, "caller err")
	}

	return &types.GetUserInfoResp{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		GetUserInfoData: struct {
			Id                string `json:"id"`
			GivenName         string `json:"given_name"`
			MiddleName        string `json:"middle_name"`
			FamilyName        string `json:"family_name"`
			Avatar            string `json:"avatar"`
			BirthdayYear      int64  `json:"birthday_year"`
			BirthdayMonth     int64  `json:"birthday_month"`
			BirthdayDay       int64  `json:"birthday_day"`
			NotificationEmail string `json:"notification_email"`
			Locate            string `json:"locate"`
			Language          string `json:"language"`
			CreateTime        string `json:"create_time"`
			LastLoginTime     string `json:"last_login_time"`
		}{
			Id:                strconv.FormatInt(user.Id, 10),
			GivenName:         user.GivenName,
			MiddleName:        user.MiddleName,
			FamilyName:        user.FamilyName,
			Avatar:            user.Avatar.String,
			BirthdayYear:      user.BirthdayYear.Int64,
			BirthdayMonth:     user.BirthdayMonth.Int64,
			BirthdayDay:       user.BirthdayDay.Int64,
			NotificationEmail: user.NotificationEmail.String,
			Locate:            user.Locate,
			Language:          user.Language,
			CreateTime:        user.CreateTime.String(),
			LastLoginTime:     user.LastLoginTime.String(),
		},
	}, nil
}
