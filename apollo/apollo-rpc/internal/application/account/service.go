// Package account orchestrates account registration, password login and
// account-profile queries.
package account

import (
	"context"
	"errors"
	"time"

	"jian-unified-system/apollo/apollo-rpc/internal/application"
	domainaccount "jian-unified-system/apollo/apollo-rpc/internal/domain/account"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/event"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
)

// AccountStore is the account-management port implemented by the MySQL
// identity adapter. Commands stay narrow so concurrent edits to unrelated
// profile fields do not overwrite each other.
type AccountStore interface {
	User(context.Context, int64) (identity.UserInfo, error)
	AuthVersion(context.Context, int64) (int64, error)
	UpdateName(context.Context, domainaccount.Profile) error
	UpdateBirthday(context.Context, domainaccount.Profile) error
	UpdateLanguage(context.Context, domainaccount.Profile) error
	PasswordHash(context.Context, int64) (string, error)
	UpdatePassword(context.Context, int64, string, string, time.Time, bool) error
	SetNotificationEmail(context.Context, int64, *string) error
	AddContact(context.Context, int64, identity.Contact) error
	Contact(context.Context, int64, int64) (identity.Contact, error)
	DeleteContact(context.Context, int64, int64) error
	DeleteAccount(context.Context, int64, string) error
}

// Service owns account registration, password login and profile-query use cases.
type Service struct {
	repository domainaccount.Repository
	passwords  domainaccount.Passwords
	login      domainaccount.LoginService
	store      AccountStore
	publisher  event.Publisher
	now        func() time.Time
}

func NewService(repository domainaccount.Repository, passwords domainaccount.Passwords, store AccountStore, publisher event.Publisher) *Service {
	if publisher == nil {
		publisher = event.NoopPublisher{}
	}
	return &Service{
		repository: repository,
		passwords:  passwords,
		login:      domainaccount.NewLoginService(passwords),
		store:      store,
		publisher:  publisher,
		now:        time.Now,
	}
}

// Registration is the input to Register.
type Registration struct {
	ID                                int64
	Email, Password, Locale, Language string
}

func (s *Service) Register(ctx context.Context, in Registration) error {
	in.Locale = domainaccount.NormalizeLocale(in.Locale)
	email, err := domainaccount.ParseEmail(in.Email)
	if err != nil {
		return err
	}
	if err = domainaccount.ValidatePassword(in.Password); err != nil {
		return err
	}
	hash, err := s.passwords.Hash(in.Password)
	if err != nil {
		return err
	}
	created, err := domainaccount.Register(in.ID, email, hash, in.Locale, in.Language)
	if err != nil {
		return err
	}
	if err = s.repository.Insert(ctx, created); err != nil {
		return err
	}
	s.publisher.Publish(event.AccountRegistered{
		Base:      event.NewBase(s.now().UTC()),
		AccountID: created.Profile().ID(),
		Email:     string(email),
	})
	return nil
}

func (s *Service) Login(ctx context.Context, address, password string) (domainaccount.Profile, error) {
	email, err := domainaccount.ParseEmail(address)
	if err != nil || password == "" || len(password) > 72 {
		return domainaccount.Profile{}, domainaccount.ErrCredentials
	}
	found, err := s.repository.FindByEmail(ctx, email)
	if errors.Is(err, application.ErrNotFound) {
		return s.login.Authenticate(nil, password)
	}
	if err != nil {
		return domainaccount.Profile{}, err
	}
	return s.login.Authenticate(&found, password)
}

func (s *Service) User(ctx context.Context, id int64) (identity.UserInfo, error) {
	if id <= 0 {
		return identity.UserInfo{}, identity.ErrInvalid
	}
	if s.store == nil {
		return identity.UserInfo{}, application.ErrNotFound
	}
	return s.store.User(ctx, id)
}

func (s *Service) SessionVersion(ctx context.Context, id int64) (int64, error) {
	if id <= 0 || s.store == nil {
		return 0, identity.ErrInvalid
	}
	return s.store.AuthVersion(ctx, id)
}

func (s *Service) UpdateName(ctx context.Context, id int64, given, middle, family string) error {
	user, err := s.User(ctx, id)
	if err != nil {
		return err
	}
	profile, err := user.Profile().WithValidatedName(given, middle, family)
	if err != nil {
		return err
	}
	if err = s.store.UpdateName(ctx, profile); err != nil {
		return err
	}
	s.publishProfile(id, "name")
	return nil
}

func (s *Service) UpdateBirthday(ctx context.Context, id, year, month, day int64) error {
	if year == 0 || month == 0 || day == 0 {
		return domainaccount.ErrInvalid
	}
	user, err := s.User(ctx, id)
	if err != nil {
		return err
	}
	profile, err := user.Profile().WithBirthday(year, month, day)
	if err != nil {
		return err
	}
	birthday := time.Date(int(year), time.Month(month), int(day), 0, 0, 0, 0, time.UTC)
	if birthday.After(s.now().UTC()) {
		return domainaccount.ErrInvalid
	}
	if err = s.store.UpdateBirthday(ctx, profile); err != nil {
		return err
	}
	s.publishProfile(id, "birthday")
	return nil
}

func (s *Service) UpdateLanguage(ctx context.Context, id int64, language string) error {
	user, err := s.User(ctx, id)
	if err != nil {
		return err
	}
	profile, err := user.Profile().WithLanguage(language)
	if err != nil {
		return err
	}
	if err = s.store.UpdateLanguage(ctx, profile); err != nil {
		return err
	}
	s.publishProfile(id, "language")
	return nil
}

func (s *Service) AddContact(ctx context.Context, id int64, value string, contactType int64, phoneRegion string) (identity.Contact, error) {
	if _, err := s.User(ctx, id); err != nil {
		return identity.Contact{}, err
	}
	var contactID int64
	for contactID <= 0 || contactID == id {
		var err error
		contactID, err = application.RandomID()
		if err != nil {
			return identity.Contact{}, err
		}
	}
	contact, err := identity.NewContact(contactID, value, contactType, phoneRegion)
	if err != nil {
		return identity.Contact{}, err
	}
	if err = s.store.AddContact(ctx, id, contact); err != nil {
		return identity.Contact{}, err
	}
	s.publisher.Publish(event.ContactChanged{Base: event.NewBase(s.now().UTC()), AccountID: id, ContactID: contact.ID(), Action: "added"})
	return contact, nil
}

func (s *Service) RemoveContact(ctx context.Context, id, contactID int64) error {
	if id <= 0 || contactID <= 0 {
		return identity.ErrInvalid
	}
	contact, err := s.store.Contact(ctx, id, contactID)
	if err != nil {
		return err
	}
	if err = contact.IsRemovable(); err != nil {
		return err
	}
	if err = s.store.DeleteContact(ctx, id, contactID); err != nil {
		return err
	}
	s.publisher.Publish(event.ContactChanged{Base: event.NewBase(s.now().UTC()), AccountID: id, ContactID: contactID, Action: "removed"})
	return nil
}

func (s *Service) ChangePassword(ctx context.Context, id int64, current, next string, signOutEverywhere bool) error {
	if id <= 0 || current == "" || len(current) > 72 || domainaccount.ValidatePassword(next) != nil {
		return domainaccount.ErrInvalid
	}
	hash, err := s.store.PasswordHash(ctx, id)
	if err != nil {
		return err
	}
	if hash == "" || !s.passwords.Verify(current, hash) {
		return application.ErrCredentials
	}
	if s.passwords.Verify(next, hash) {
		return domainaccount.ErrInvalid
	}
	nextHash, err := s.passwords.Hash(next)
	if err != nil {
		return err
	}
	now := s.now().UTC()
	if err = s.store.UpdatePassword(ctx, id, hash, nextHash, now, signOutEverywhere); err != nil {
		return err
	}
	s.publisher.Publish(event.PasswordChanged{Base: event.NewBase(now), AccountID: id})
	return nil
}

func (s *Service) ChangeNotificationEmail(ctx context.Context, id int64, value string) error {
	if id <= 0 {
		return domainaccount.ErrInvalid
	}
	if _, err := domainaccount.ParseEmail(value); err != nil {
		return err
	}
	if err := s.store.SetNotificationEmail(ctx, id, &value); err != nil {
		return err
	}
	s.publisher.Publish(event.NotificationEmailChanged{Base: event.NewBase(s.now().UTC()), AccountID: id})
	return nil
}

func (s *Service) RemoveNotificationEmail(ctx context.Context, id int64) error {
	if id <= 0 {
		return domainaccount.ErrInvalid
	}
	if err := s.store.SetNotificationEmail(ctx, id, nil); err != nil {
		return err
	}
	s.publisher.Publish(event.NotificationEmailChanged{Base: event.NewBase(s.now().UTC()), AccountID: id, Removed: true})
	return nil
}

func (s *Service) DeleteAccount(ctx context.Context, id int64, currentPassword, confirmation string) error {
	if id <= 0 {
		return domainaccount.ErrInvalid
	}
	if err := domainaccount.ConfirmDeletion(confirmation); err != nil {
		return err
	}
	hash, err := s.store.PasswordHash(ctx, id)
	if err != nil {
		return err
	}
	if hash != "" && (currentPassword == "" || len(currentPassword) > 72 || !s.passwords.Verify(currentPassword, hash)) {
		return application.ErrCredentials
	}
	if err = s.store.DeleteAccount(ctx, id, hash); err != nil {
		return err
	}
	s.publisher.Publish(event.AccountDeleted{Base: event.NewBase(s.now().UTC()), AccountID: id})
	return nil
}

func (s *Service) publishProfile(id int64, field string) {
	s.publisher.Publish(event.AccountProfileUpdated{Base: event.NewBase(s.now().UTC()), AccountID: id, Field: field})
}
