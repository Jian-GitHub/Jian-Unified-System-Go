package oauth2

import (
	"errors"
	"strconv"
	"strings"
)

type GoogleUserProfile struct {
	ID           int64
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

func (g GoogleUserProfile) SetGoogleID() error {
	// 1. Check Google ID
	parts := strings.Split(g.ResourceName, "/")
	if len(parts) < 2 {
		return errors.New("not enough parts")
	}
	// 2. Set Google ID -> GoogleUserProfile
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return err
	}
	g.ID = id
	return nil
}
