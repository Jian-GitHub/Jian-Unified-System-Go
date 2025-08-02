package githubUtil

import (
	"database/sql"
	"jian-unified-system/apollo/apollo-rpc/internal/model"
	"jian-unified-system/jus-core/types/oauth2"
	"strings"
)

func ParseGitHubUserInfo(userInfo *oauth2.GitHubUserInfo) (user *model.User, contacts []*model.Contact, err error) {
	// 1. Set Name
	if userInfo.Name == nil || *userInfo.Name == "" {
		user.GivenName = userInfo.Login
	} else {
		parts := strings.Split(*userInfo.Name, " ")
		switch len(parts) {
		case 0:
			user.GivenName = userInfo.Login
		case 1:
			user.GivenName = parts[0]
		case 2:
			user.GivenName = parts[0]
			user.FamilyName = parts[1]
		case 3:
			user.GivenName = parts[0]
			user.MiddleName = parts[1]
			user.FamilyName = parts[2]
		}
	}

	// 2. Set Email
	if userInfo.Email != nil {
		user.Email = *userInfo.Email
	}
	// 3. Set NotificationEmail
	if userInfo.NotificationEmail != nil {
		user.NotificationEmail = sql.NullString{
			String: *userInfo.NotificationEmail,
			Valid:  true,
		}
	}
	// 4. Set Avatar
	user.Avatar = sql.NullString{
		String: userInfo.AvatarURL,
		Valid:  true,
	}

	return user, nil, nil
}
