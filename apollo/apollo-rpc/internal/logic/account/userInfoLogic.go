package accountlogic

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"
)

type UserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserInfoLogic {
	return &UserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UserInfo 用户信息
func (l *UserInfoLogic) UserInfo(in *apollo.UserInfoReq) (*apollo.UserInfoResp, error) {
	user, err := l.svcCtx.UserModel.FindOneUserInfo(l.ctx, in.UserId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("no user")
	}

	//if !user.NotificationEmail.Valid {
	//	return &apollo.UserInfoResp{
	//		UserBytes: nil,
	//	}, nil
	//}
	if user.NotificationEmail.Valid {
		email, err := l.svcCtx.MLKEMKeyManager.DecryptMessage(user.NotificationEmail.String)
		if err != nil {
			return nil, err
		}

		user.NotificationEmail.String = email
	}

	userBytes, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	return &apollo.UserInfoResp{
		UserBytes: userBytes,
	}, nil
}
