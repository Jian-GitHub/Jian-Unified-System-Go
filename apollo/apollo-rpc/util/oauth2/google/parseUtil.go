package googleUtil

import (
	"database/sql"
	"jian-unified-system/apollo/apollo-rpc/internal/model"
	"jian-unified-system/jus-core/types/oauth2"
	sqlType "jian-unified-system/jus-core/types/sql"
)

func ParseUserProfile(userProfile *oauth2.GoogleUserProfile) (user *model.User, contacts []*model.Contact, err error) {
	// 3. Parse SQL User
	err = parseUser(userProfile, user)
	if err != nil {
		return nil, nil, err
	}

	// 4. Parse SQL Contacts (Email)
	err = parseContact(userProfile, user, contacts)
	if err != nil {
		return nil, nil, err
	}

	return user, contacts, nil
}

func parseUser(userProfile *oauth2.GoogleUserProfile, user *model.User) error {
	// 1. 主要名字
	for _, name := range userProfile.Names {
		if name.Metadata.Primary {
			user.GivenName = name.GivenName
			user.FamilyName = name.FamilyName
			break
		}
	}

	// 2. 主要照片
	for _, photo := range userProfile.Photos {
		if photo.Metadata.Primary {
			user.Avatar = sql.NullString{
				String: photo.URL,
				Valid:  true,
			}
			break
		}
	}

	// 3. 主要生日
	for _, birthday := range userProfile.Birthdays {
		if birthday.Metadata.Primary {
			user.BirthdayYear = sql.NullInt64{
				Int64: birthday.Date.Year,
				Valid: true,
			}
			user.BirthdayMonth = sql.NullInt64{
				Int64: birthday.Date.Month,
				Valid: true,
			}
			user.BirthdayDay = sql.NullInt64{
				Int64: birthday.Date.Day,
				Valid: true,
			}
			break
		}
	}
	return nil
}

func parseContact(userProfile *oauth2.GoogleUserProfile, user *model.User, contacts []*model.Contact) error {
	// 4. 已验证邮件
	for _, email := range userProfile.EmailAddresses {
		if email.Metadata.Verified {
			if email.Metadata.Primary {
				(*user).Email = email.Value
				(*user).NotificationEmail = sql.NullString{
					String: email.Value,
					Valid:  true,
				}
			}
			contacts = append(contacts, &model.Contact{
				UserId: userProfile.ID,
				Value:  email.Value,
				Type:   sqlType.ContactType.Email,
			})
		}
	}
	return nil
}
