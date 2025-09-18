package accountlogic

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/apollo/apollo-rpc/internal/svc"
	ap "jian-unified-system/jus-core/data/mysql/apollo"
	"jian-unified-system/jus-core/types/oauth2"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserSecurityInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserSecurityInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserSecurityInfoLogic {
	return &UserSecurityInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UserSecurityInfo 用户安全信息
func (l *UserSecurityInfoLogic) UserSecurityInfo(in *apollo.UserSecurityInfoReq) (*apollo.UserSecurityInfoResp, error) {
	var (
		contacts *[]ap.Contact
		//passwordUpdateTime *ap.UserPasswordUpdateTime
		user         *ap.User
		tokenNum     int64
		passkeysNum  int64
		thirdParties *[]ap.ThirdParty
	)

	ctx, cancel := context.WithTimeout(l.ctx, 10*time.Second)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	// contacts
	g.Go(func() error {
		var err error
		contacts, err = l.svcCtx.ContactModel.FindByUserID(ctx, in.UserId)
		return err
	})
	// password update time
	g.Go(func() error {
		var err error
		//passwordUpdateTime, err = l.svcCtx.UserModel.FindPasswordUpdateTime(ctx, in.UserId)
		user, err = l.svcCtx.UserModel.FindOne(ctx, in.UserId)
		return err
	})
	// count tokens
	g.Go(func() error {
		var err error
		tokenNum, err = l.svcCtx.TokenModel.CountTokens(ctx, in.UserId)
		return err
	})
	// count passkeys
	g.Go(func() error {
		var err error
		passkeysNum, err = l.svcCtx.PasskeyModel.CountPasskeys(ctx, in.UserId)
		return err
	})
	// third parties
	g.Go(func() error {
		var err error
		thirdParties, err = l.svcCtx.ThirdPartyModel.FindBatch(ctx, in.UserId)
		return err
	})

	// 等待所有并发任务完成
	if err := g.Wait(); err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	// 整理 contacts
	var securityContacts []*apollo.UserContact
	if contacts != nil {
		securityContacts = make([]*apollo.UserContact, 0, len(*contacts))
		for _, contact := range *contacts {
			securityContacts = append(securityContacts, &apollo.UserContact{
				Id:          strconv.FormatInt(contact.Id, 10),
				Value:       contact.Value,
				Type:        contact.Type,
				PhoneRegion: contact.PhoneRegion,
			})
		}
	} else {
		securityContacts = make([]*apollo.UserContact, 0)
	}

	// 整理 password update date
	var updatedDate *apollo.PasswordUpdatedDate
	if user != nil && user.PasswordUpdateTime.Valid {
		updatedDate = &apollo.PasswordUpdatedDate{
			Year:  int64(user.PasswordUpdateTime.Time.Year()),
			Month: int64(user.PasswordUpdateTime.Time.Month()),
			Day:   int64(user.PasswordUpdateTime.Time.Day()),
		}
	}
	// 整理 third party accounts
	var thirdPartyAccounts = &apollo.ThirdPartyAccounts{}
	if thirdParties != nil {
		for _, tp := range *thirdParties {
			switch tp.Provider {
			case oauth2.ProviderGithub:
				thirdPartyAccounts.Github = true
			case oauth2.ProviderGoogle:
				thirdPartyAccounts.Google = true
			}
		}
	}

	//// contacts
	//contacts, err := l.svcCtx.ContactModel.FindByUserID(l.ctx, in.UserId)
	//if err != nil {
	//	return nil, err
	//}
	//securityContacts := make([]*apollo.UserContact, 0, len(contacts))
	//for index, contact := range contacts {
	//	securityContacts[index] = &apollo.UserContact{
	//		Id:          strconv.FormatInt(contact.Id, 10),
	//		Value:       contact.Value,
	//		Type:        contact.Type,
	//		PhoneRegion: contact.PhoneRegion,
	//	}
	//}
	//
	//// pwd updated time
	//var updatedDate *apollo.PasswordUpdatedDate
	//passwordUpdateTime, err = l.svcCtx.UserModel.FindPasswordUpdateTime(l.ctx, in.UserId)
	//if err != nil {
	//	return nil, err
	//} else if passwordUpdateTime != nil {
	//	updatedDate = &apollo.PasswordUpdatedDate{
	//		Year: int64(passwordUpdateTime.PasswordUpdateTime.Year()),
	//	}
	//}
	//
	//// AccountSecurityTokenNum
	//tokenNum, err := l.svcCtx.TokenModel.CountTokens(l.ctx, in.UserId)
	//if err != nil {
	//	return nil, err
	//}
	//
	//// PasskeysNum
	//passkeysNum, err := l.svcCtx.PasskeyModel.CountPasskeys(l.ctx, in.UserId)
	//if err != nil {
	//	return nil, err
	//}
	//
	//// ThirdPartyAccounts
	//var thirdPartyAccounts *apollo.ThirdPartyAccounts
	//thirdParties, err := l.svcCtx.ThirdPartyModel.FindBatch(l.ctx, in.UserId)
	//if err != nil {
	//	return nil, err
	//} else if thirdParties != nil {
	//	for _, thirdParty := range thirdParties {
	//		switch thirdParty.Provider {
	//		case oauth2.ProviderGithub:
	//			thirdPartyAccounts.Github = true
	//		case oauth2.ProviderGoogle:
	//			thirdPartyAccounts.Google = true
	//		}
	//	}
	//}

	return &apollo.UserSecurityInfoResp{
		Contacts:                securityContacts,
		PasswordUpdatedDate:     updatedDate,
		AccountSecurityTokenNum: tokenNum,
		PasskeysNum:             passkeysNum,
		ThirdPartyAccounts:      thirdPartyAccounts,
	}, nil
}
