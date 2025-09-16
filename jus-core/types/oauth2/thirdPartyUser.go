package oauth2

import "database/sql"

// ThirdPartyUser 统一接口
type ThirdPartyUser interface {
	GetID() string
	GetProvider() string
	GetGivenName() string
	GetMiddleName() string
	GetFamilyName() string
	GetEmail() string
	GetNotificationEmail() sql.NullString
	GetAvatar() sql.NullString
	GetBirthdayYear() sql.NullInt64
	GetBirthdayMonth() sql.NullInt64
	GetBirthdayDay() sql.NullInt64
	GetDisplayName() string
	GenerateEmailContacts() *[][2]interface{}
}
