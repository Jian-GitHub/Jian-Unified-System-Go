package accountlogic

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"
)

type NotificationInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewNotificationInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NotificationInfoLogic {
	return &NotificationInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// NotificationInfo 用户信息
func (l *NotificationInfoLogic) NotificationInfo(in *apollo.NotificationInfoReq) (*apollo.NotificationInfoResp, error) {
	userNotificationInfo, err := l.svcCtx.UserModel.FindOneNotificationInfo(l.ctx, in.UserId)
	if err != nil {
		return nil, err
	}
	if userNotificationInfo == nil {
		return nil, errors.New("no user")
	}

	if !userNotificationInfo.NotificationEmail.Valid {
		return &apollo.NotificationInfoResp{
			UserBytes: nil,
		}, nil
	}

	email, err := l.svcCtx.MLKEMKeyManager.DecryptMessage(userNotificationInfo.NotificationEmail.String)
	if err != nil {
		return nil, err
	}

	userNotificationInfo.NotificationEmail.String = email

	userBytes, err := json.Marshal(userNotificationInfo)
	if err != nil {
		return nil, err
	}
	return &apollo.NotificationInfoResp{
		UserBytes: userBytes,
	}, nil
}
