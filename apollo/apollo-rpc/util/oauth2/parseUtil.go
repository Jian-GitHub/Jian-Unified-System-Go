package oauth2

import (
	"jian-unified-system/apollo/apollo-rpc/internal/model"
	"jian-unified-system/jus-core/types/oauth2"
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

	// generate contacts
	contactsData := (*thirdPartyUser).GenerateEmailContacts()
	for _, data := range *contactsData {
		*contacts = append(*contacts, &model.Contact{
			UserId:    user.Id,
			Value:     data[0].(string),
			Type:      data[1].(int64),
			IsEnabled: data[2].(int64),
		})
	}

	return nil
}
