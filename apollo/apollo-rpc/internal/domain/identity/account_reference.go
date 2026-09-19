package identity

import "encoding/json"

// AccountReference is the anti-corruption value passed from the account
// context into identity use cases. It carries only the stable account key and
// localization hints required by authentication ceremonies.
type AccountReference struct {
	id       int64
	locale   string
	language string
}

func NewAccountReference(id int64, locale, language string) (AccountReference, error) {
	if id <= 0 || locale == "" || language == "" || len(locale) > 32 || len(language) > 32 {
		return AccountReference{}, ErrInvalid
	}
	return AccountReference{id: id, locale: locale, language: language}, nil
}

func (r AccountReference) ID() int64        { return r.id }
func (r AccountReference) Locale() string   { return r.locale }
func (r AccountReference) Language() string { return r.language }

func (r AccountReference) WithLocale(locale, language string) (AccountReference, error) {
	return NewAccountReference(r.id, locale, language)
}

func (r AccountReference) Valid() bool {
	_, err := NewAccountReference(r.id, r.locale, r.language)
	return err == nil
}

// MarshalJSON preserves the authentication-session wire shape while the value
// object's fields remain private.
func (r AccountReference) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID       int64  `json:"id"`
		Locale   string `json:"locale"`
		Language string `json:"language"`
	}{ID: r.id, Locale: r.locale, Language: r.language})
}

func (r *AccountReference) UnmarshalJSON(body []byte) error {
	var raw struct {
		ID       int64  `json:"id"`
		Locale   string `json:"locale"`
		Language string `json:"language"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return err
	}
	parsed, err := NewAccountReference(raw.ID, raw.Locale, raw.Language)
	if err != nil {
		return err
	}
	*r = parsed
	return nil
}
