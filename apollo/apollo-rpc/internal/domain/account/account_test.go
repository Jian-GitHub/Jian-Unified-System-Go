package account

import (
	"strings"
	"testing"
	"time"
)

func TestRegistrationRules(t *testing.T) {
	for _, value := range []string{"", "plain", "missing@host", " user@example.com", strings.Repeat("a", 250) + "@example.com"} {
		if _, err := ParseEmail(value); err == nil {
			t.Errorf("accepted invalid email %q", value)
		}
	}
	e, err := ParseEmail("User@example.com")
	if err != nil || string(e) != "User@example.com" {
		t.Fatal("email lookup semantics changed")
	}
	for _, value := range []string{"", "Ab1!", "abcdefgh1!", "ABCDEFGH1!", "Abcdefgh!", "Abcdefgh1", "Abc1!" + strings.Repeat("x", 68)} {
		if ValidatePassword(value) == nil {
			t.Errorf("accepted invalid password of length %d", len(value))
		}
	}
	if err := ValidatePassword("Valid123!"); err != nil {
		t.Fatal(err)
	}
	if _, err := Register(0, e, "hash", "CN", "zh"); err == nil {
		t.Fatal("accepted non-positive ID")
	}
	a, err := Register(1, e, "hash", "CN", "zh")
	if err != nil {
		t.Fatal(err)
	}
	if a.Profile().ID() != 1 {
		t.Fatal("profile did not carry the registered id")
	}
	// Profiles are value objects; constructing a new one with a
	// different id must not bleed into the aggregate.
	replacement, err := NewProfile(2, "", "", "", "", "CN", "zh", 0, 0, 0, time.Time{}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if replacement.ID() == a.Profile().ID() {
		t.Fatal("profile accessor returned the original id (immutability check)")
	}
}

func TestNormalizeLocale(t *testing.T) {
	if NormalizeLocale("") != UnknownLocale || NormalizeLocale("NZ") != "NZ" {
		t.Fatal("locale normalization changed")
	}
}

func TestAccountAggregateUpdates(t *testing.T) {
	a, err := Register(1, "user@example.com", "hash-v1", "CN", "zh")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	rotated, err := a.UpdatePassword("hash-v2", now)
	if err != nil {
		t.Fatal(err)
	}
	if rotated.PasswordHash() != "hash-v2" || a.PasswordHash() != "hash-v1" {
		t.Fatal("original aggregate must stay immutable")
	}
	if rotated.Profile().PasswordUpdatedAt().IsZero() {
		t.Fatal("password rotation must stamp PasswordUpdatedAt")
	}
	if _, err := rotated.UpdatePassword("", now); err == nil {
		t.Fatal("empty hash accepted")
	}
	if _, err := rotated.UpdatePassword("hash-v3", time.Time{}); err == nil {
		t.Fatal("zero timestamp accepted")
	}
	localised, err := rotated.ChangeLocale("US", "en")
	if err != nil {
		t.Fatal(err)
	}
	if localised.Profile().Locale() != "US" || localised.Profile().Language() != "en" {
		t.Fatal("locale change did not apply")
	}
	if rotated.Profile().Locale() != "CN" {
		t.Fatal("locale change leaked into original aggregate")
	}
	if _, err := localised.ChangeLocale("", "en"); err == nil {
		t.Fatal("empty locale accepted")
	}
}

func TestProfileValueObject(t *testing.T) {
	now := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)
	p, err := NewProfile(1, "Alice", "Q", "Tester", "https://cdn/avatar.png", "CN", "zh", 1990, 1, 2, now, 3)
	if err != nil {
		t.Fatal(err)
	}
	if p.ID() != 1 || p.GivenName() != "Alice" || p.Avatar() != "https://cdn/avatar.png" || p.BirthdayYear() != 1990 || !p.PasswordUpdatedAt().Equal(now) {
		t.Fatal("constructor did not preserve fields")
	}
	if _, err := NewProfile(0, "", "", "", "", "CN", "zh", 0, 0, 0, time.Time{}, 0); err == nil {
		t.Fatal("zero id accepted")
	}
	if _, err := NewProfile(1, "", "", "", "", "", "zh", 0, 0, 0, time.Time{}, 0); err == nil {
		t.Fatal("empty locale accepted")
	}
	named := p.WithName("Bob", "", "")
	if named.GivenName() != "Bob" || p.GivenName() != "Alice" {
		t.Fatal("WithName leaked into original")
	}
	avatar := p.WithAvatar("https://cdn/avatar2.png")
	if avatar.Avatar() != "https://cdn/avatar2.png" || p.Avatar() != "https://cdn/avatar.png" {
		t.Fatal("WithAvatar leaked into original")
	}
	birthday, err := p.WithBirthday(2000, 5, 20)
	if err != nil {
		t.Fatal(err)
	}
	if birthday.BirthdayYear() != 2000 || birthday.BirthdayMonth() != 5 || birthday.BirthdayDay() != 20 {
		t.Fatal("birthday update lost")
	}
	for _, birthday := range [][3]int64{{-1, 1, 1}, {2000, 13, 1}, {2000, 1, 32}, {2025, 2, 29}, {2024, 2, 30}} {
		if _, err := p.WithBirthday(birthday[0], birthday[1], birthday[2]); err == nil {
			t.Fatalf("invalid birthday accepted: %v", birthday)
		}
	}
	named, err = p.WithValidatedName("Bob", "Q", "Tester")
	if err != nil || named.FamilyName() != "Tester" {
		t.Fatal("validated name update failed")
	}
	for _, names := range [][3]string{{"", "", "Tester"}, {" Bob", "", "Tester"}, {"Bob", "", ""}} {
		if _, err = p.WithValidatedName(names[0], names[1], names[2]); err == nil {
			t.Fatalf("invalid name accepted: %q", names)
		}
	}
	translated, err := p.WithLanguage("ja")
	if err != nil || translated.Language() != "ja" || translated.Locale() != p.Locale() {
		t.Fatal("language update changed the locale")
	}
	if _, err = p.WithLanguage(" ja"); err == nil {
		t.Fatal("untrimmed language accepted")
	}
	if _, err := Restore("u@example.com", "hash", Profile{}); err == nil {
		t.Fatal("Restore accepted empty profile")
	}
	a, err := Restore("u@example.com", "hash", p)
	if err != nil {
		t.Fatal(err)
	}
	updated, err := a.UpdatePassword("hash-2", now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if updated.Profile().PasswordUpdatedAt() != now.Add(time.Hour) {
		t.Fatal("UpdatePassword did not stamp the Profile timestamp")
	}
	external := a.ApplyExternalProfile("Alice Q", "https://cdn/oauth.png")
	if external.Profile().GivenName() != "Alice Q" || external.Profile().Avatar() != "https://cdn/oauth.png" {
		t.Fatal("ApplyExternalProfile did not project values")
	}
}
