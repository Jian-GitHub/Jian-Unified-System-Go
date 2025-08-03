package oauth2

import (
	"database/sql"
	"fmt"
	"jian-unified-system/apollo/apollo-rpc/internal/model"
	"jian-unified-system/jus-core/types/oauth2"
	sqlType "jian-unified-system/jus-core/types/sql"
)

func ParseUserAndContacts(thirdPartyUser *oauth2.ThirdPartyUser, user *model.User, contacts *[]*model.Contact, thirdParty *model.ThirdParty) (err error) {
	// set user fields
	user.GivenName = (*thirdPartyUser).GetGivenName()
	user.MiddleName = (*thirdPartyUser).GetMiddleName()
	user.FamilyName = (*thirdPartyUser).GetFamilyName()
	user.Email = (*thirdPartyUser).GetEmail()
	user.EmailVerified = 1
	user.Avatar = (*thirdPartyUser).GetAvatar()
	user.BirthdayYear = (*thirdPartyUser).GetBirthdayYear()
	user.BirthdayMonth = (*thirdPartyUser).GetBirthdayMonth()
	user.BirthdayDay = (*thirdPartyUser).GetBirthdayDay()
	user.NotificationEmail = (*thirdPartyUser).GetNotificationEmail()

	// set thirdParty fields
	thirdParty.ThirdId = (*thirdPartyUser).GetID()
	thirdParty.UserId = user.Id
	thirdParty.Name = (*thirdPartyUser).GetDisplayName()

	// 准备 contacts
	switch v := (*thirdPartyUser).(type) {
	case *oauth2.GitHubAdapter:
		if v.Email != nil && *v.Email != "" {
			*contacts = append(*contacts, &model.Contact{
				UserId: user.Id,
				Value:  *v.Email,
				Type:   sqlType.ContactType.Email,
			})
		}
		if v.NotificationEmail != nil && *v.NotificationEmail != "" && *v.NotificationEmail != *v.Email {
			*contacts = append(*contacts, &model.Contact{
				UserId: user.Id,
				Value:  *v.NotificationEmail,
				Type:   sqlType.ContactType.Email,
			})
		}

	case *oauth2.GoogleAdapter:
		for _, email := range v.EmailAddresses {
			if email.Metadata.Verified {
				if email.Metadata.Primary {
					(*user).Email = email.Value
					(*user).NotificationEmail = sql.NullString{
						String: email.Value,
						Valid:  true,
					}
				}
				*contacts = append(*contacts, &model.Contact{
					UserId:    user.Id,
					Value:     email.Value,
					Type:      sqlType.ContactType.Email,
					IsEnabled: 1,
				})
			}
		}
	default:
		fmt.Println("未知类型")
		return
	}

	return nil
}
