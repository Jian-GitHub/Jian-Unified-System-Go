package accountlogic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/jsonx"
	"jian-unified-system/apollo/apollo-rpc/internal/model"
	"jian-unified-system/jus-core/util"

	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegistrationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegistrationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegistrationLogic {
	return &RegistrationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Registration 注册
func (l *RegistrationLogic) Registration(in *apollo.RegistrationReq) (*apollo.Empty, error) {
	// todo: add your logic here and delete this line
	//fmt.Println("enter Reg RPC")
	email := util.HashSHA512(in.Email, in.Email)
	// Check if the email exists
	_, err := l.svcCtx.UserModel.FindOneByEmail(l.ctx, email)
	if err == nil {
		return nil, errors.New("user exists")
	}

	// All params are validated
	// Generate new user
	pwd, err := util.HashPasswordBcrypt(in.Password)
	if err != nil {
		return nil, err
	}
	//encryptedEmail, err := l.svcCtx.MLKEMKeyManager.EncryptMessage(in.Email)
	//if err != nil {
	//	return nil, err
	//}
	//fmt.Println("encryptedEmail")
	//fmt.Println(encryptedEmail)
	//fmt.Println("decryptedEmail")
	//
	//decryptedEmail, err := l.svcCtx.MLKEMKeyManager.DecryptMessage(encryptedEmail)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return nil, err
	//}
	//fmt.Println(decryptedEmail)

	NotificationEmail, err := l.svcCtx.MLKEMKeyManager.EncryptMessage(in.Email)
	if err != nil {
		return nil, err
	}

	//fmt.Println(email)
	user := &model.User{
		Id:       in.UserId,
		Email:    email,
		Password: pwd,
		Locate:   in.Locate,
		Language: in.Language,
		NotificationEmail: sql.NullString{
			String: NotificationEmail,
			Valid:  NotificationEmail != "",
		},
	}

	//user, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return nil, err
	//}
	_, err = l.svcCtx.UserModel.Insert(l.ctx, user)
	if err != nil {
		return nil, err
	}
	toString, err := jsonx.MarshalToString(user)
	if err != nil {
		return nil, err
	}
	fmt.Println(toString)
	return &apollo.Empty{}, nil
}
