package oauth2

import (
	"jian-unified-system/jus-core/data/mysql/apollo"
	"jian-unified-system/jus-core/types/oauth2"
)

func ParseUserAndContacts(thirdPartyUser *oauth2.ThirdPartyUser, user *apollo.User, contacts *[]*apollo.Contact, thirdParty *apollo.ThirdParty) (err error) {
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
	thirdParty.Provider = (*thirdPartyUser).GetProvider()
	thirdParty.UserId = user.Id
	thirdParty.Name = (*thirdPartyUser).GetDisplayName()

	// generate contacts
	if contacts == nil {
		return nil
	}

	contactsData := (*thirdPartyUser).GenerateEmailContacts()
	for _, data := range *contactsData {
		*contacts = append(*contacts, &apollo.Contact{
			UserId:    user.Id,
			Value:     data[0].(string),
			Type:      data[1].(int64),
			IsEnabled: 1,
		})
	}

	return nil
}
