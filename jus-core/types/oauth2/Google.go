package oauth2

import (
	"database/sql"
	sqlType "jian-unified-system/jus-core/types/sql"
	"strings"
)

type GoogleUserProfile struct {
	ID           string
	ResourceName string `json:"resourceName"`
	Etag         string `json:"etag"`
	Names        []struct {
		Metadata struct {
			Primary       bool `json:"primary"`
			SourcePrimary bool `json:"sourcePrimary,omitempty"`
			Source        struct {
				Type string `json:"type"`
				ID   string `json:"id"`
			} `json:"source"`
		} `json:"metadata"`
		DisplayName          string `json:"displayName"`
		FamilyName           string `json:"familyName,omitempty"`
		GivenName            string `json:"givenName,omitempty"`
		DisplayNameLastFirst string `json:"displayNameLastFirst,omitempty"`
		UnstructuredName     string `json:"unstructuredName,omitempty"`
	} `json:"names"`
	Photos []struct {
		Metadata struct {
			Primary bool `json:"primary"`
			Source  struct {
				Type string `json:"type"`
				ID   string `json:"id"`
			} `json:"source"`
		} `json:"metadata"`
		URL     string `json:"url"`
		Default bool   `json:"default,omitempty"`
	} `json:"photos"`
	Birthdays []struct {
		Metadata struct {
			Primary bool `json:"primary"`
			Source  struct {
				Type string `json:"type"`
				ID   string `json:"id"`
			} `json:"source"`
		} `json:"metadata"`
		Date struct {
			Year  int64 `json:"year"`
			Month int64 `json:"month"`
			Day   int64 `json:"day"`
		} `json:"date"`
	} `json:"birthdays"`
	EmailAddresses []struct {
		Metadata struct {
			Primary       bool `json:"primary,omitempty"`
			Verified      bool `json:"verified"`
			SourcePrimary bool `json:"sourcePrimary,omitempty"`
			Source        struct {
				Type string `json:"type"`
				ID   string `json:"id"`
			} `json:"source"`
		} `json:"metadata"`
		Value string `json:"value"`
	} `json:"emailAddresses"`
	Locales []struct {
		Metadata struct {
			Primary bool `json:"primary"`
			Source  struct {
				Type string `json:"type"`
				ID   string `json:"id"`
			} `json:"source"`
		} `json:"metadata"`
		Value string `json:"value"`
	} `json:"locales,omitempty"`
	PhoneNumbers []struct {
		Metadata struct {
			Primary bool `json:"primary"`
			Source  struct {
				Type string `json:"type"`
				ID   string `json:"id"`
			} `json:"source"`
		} `json:"metadata"`
		Value         string `json:"value"`
		CanonicalForm string `json:"canonicalForm,omitempty"`
	} `json:"phoneNumbers,omitempty"`
	// 其他可能存在的字段
	Metadata struct {
		Sources []struct {
			Type string `json:"type"`
			ID   string `json:"id"`
		} `json:"sources,omitempty"`
	} `json:"metadata,omitempty"`
}

type GoogleAdapter struct {
	*GoogleUserProfile
}

//func (g GoogleAdapter) SetGoogleID() error {
//	// 1. Check Google ID
//	parts := strings.Split(g.ResourceName, "/")
//	if len(parts) < 2 {
//		return errors.New("not enough parts")
//	}
//	// 2. Set Google ID -> GoogleUserProfile
//	id, err := strconv.ParseInt(parts[1], 10, 64)
//	if err != nil {
//		return err
//	}
//	g.ID = id
//	return nil
//}

func (g GoogleAdapter) GenerateEmailContacts() *[][3]interface{} {
	contacts := make([][3]interface{}, 0)
	//notificationEmail := sql.NullString{
	//	String: "",
	//	Valid:  false,
	//}

	for _, email := range g.EmailAddresses {
		if email.Metadata.Verified {
			//if email.Metadata.Primary {
			//notificationEmail.String = email.Value
			//notificationEmail.Valid = true

			//(*user).Email = email.Value
			//(*user).NotificationEmail = sql.NullString{
			//	String: email.Value,
			//	Valid:  true,
			//}
			//}
			contacts = append(contacts, [3]interface{}{email.Value, sqlType.ContactType.Email, 1})
		}
	}
	return &contacts //, notificationEmail
}
func (g GoogleAdapter) GetID() string {
	parts := strings.Split(g.ResourceName, "/")
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}
func (g GoogleAdapter) GetGivenName() string {
	for _, name := range g.Names {
		if name.Metadata.Primary {
			return name.GivenName
		}
	}
	return ""
}
func (g GoogleAdapter) GetMiddleName() string {
	return ""
}
func (g GoogleAdapter) GetFamilyName() string {
	for _, name := range g.Names {
		if name.Metadata.Primary {
			return name.FamilyName
		}
	}
	return ""
}
func (g GoogleAdapter) GetEmail() string {
	for _, email := range g.EmailAddresses {
		if email.Metadata.Verified {
			if email.Metadata.Primary {
				return email.Value
			}
		}
	}
	return ""
}
func (g GoogleAdapter) GetNotificationEmail() sql.NullString {
	for _, email := range g.EmailAddresses {
		if email.Metadata.Verified {
			if email.Metadata.Primary {
				return sql.NullString{
					String: email.Value,
					Valid:  true,
				}
			}
		}
	}
	return sql.NullString{Valid: false}
}
func (g GoogleAdapter) GetBirthdayYear() sql.NullInt64 {
	for _, birthday := range g.Birthdays {
		if birthday.Metadata.Primary {
			return sql.NullInt64{
				Int64: birthday.Date.Year,
				Valid: true,
			}
		}
	}
	return sql.NullInt64{Valid: false}
}
func (g GoogleAdapter) GetBirthdayMonth() sql.NullInt64 {
	for _, birthday := range g.Birthdays {
		if birthday.Metadata.Primary {
			return sql.NullInt64{
				Int64: birthday.Date.Month,
				Valid: true,
			}
		}
	}
	return sql.NullInt64{Valid: false}
}
func (g GoogleAdapter) GetBirthdayDay() sql.NullInt64 {
	for _, birthday := range g.Birthdays {
		if birthday.Metadata.Primary {
			return sql.NullInt64{
				Int64: birthday.Date.Day,
				Valid: true,
			}
		}
	}
	return sql.NullInt64{Valid: false}
}
func (g GoogleAdapter) GetAvatar() sql.NullString {
	for _, photo := range g.Photos {
		if photo.Metadata.Primary {
			return sql.NullString{
				String: photo.URL,
				Valid:  true,
			}
		}
	}

	return sql.NullString{Valid: false}
}
func (g GoogleAdapter) GetDisplayName() string {
	for _, name := range g.Names {
		if name.Metadata.Primary {
			return name.DisplayName
		}
	}
	return "Google"
}
