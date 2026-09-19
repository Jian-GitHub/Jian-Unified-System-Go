package passkey

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"jian-unified-system/apollo/apollo-rpc/internal/application"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/account"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/event"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
)

// Service owns WebAuthn credential lifecycle and security-summary queries.
type Service struct {
	store       identity.Store
	verifier    identity.VerifyCeremony
	credentials identity.CredentialService
	publisher   event.Publisher
	now         func() time.Time
}

func NewService(store identity.Store, verifier identity.VerifyCeremony, publisher event.Publisher) *Service {
	if publisher == nil {
		publisher = event.NoopPublisher{}
	}
	return &Service{
		store:       store,
		verifier:    verifier,
		credentials: identity.NewCredentialService(),
		publisher:   publisher,
		now:         time.Now,
	}
}

func (s *Service) Security(ctx context.Context, id int64) (identity.SecurityInfo, error) {
	if id <= 0 {
		return identity.SecurityInfo{}, identity.ErrInvalid
	}
	return s.store.Security(ctx, id)
}

func (s *Service) List(ctx context.Context, id, page int64) ([]identity.Passkey, error) {
	if id <= 0 || !identity.Page(page) {
		return nil, identity.ErrInvalid
	}
	return s.store.Passkeys(ctx, id, page)
}

func (s *Service) Remove(ctx context.Context, id int64, key string) error {
	if id <= 0 || key == "" || len(key) > 1024 {
		return identity.ErrInvalid
	}
	found, err := s.store.Credential(ctx, key)
	if errors.Is(err, application.ErrNotFound) {
		return application.ErrNotFound
	}
	if err != nil {
		return err
	}
	if found.OwnerID() != id {
		return application.ErrNotFound
	}
	lock, err := s.store.LockAccount(ctx, found.OwnerID())
	if err != nil {
		return err
	}
	defer lock.Rollback()
	inventory, err := s.store.CountCredentials(ctx, lock)
	if err != nil {
		return err
	}
	if err = s.credentials.CanRemovePasskey(found, inventory); err != nil {
		return err
	}
	if err = s.store.DeletePasskey(ctx, lock, key); err != nil {
		return err
	}
	if err = lock.Commit(); err != nil {
		return err
	}
	s.publisher.Publish(event.CredentialRemoved{
		Base:      event.NewBase(s.now().UTC()),
		AccountID: id,
		PasskeyID: key,
	})
	return nil
}

func (s *Service) StartRegistration(ctx context.Context, profile account.Profile, name string, bind bool) ([]byte, string, error) {
	reference, err := referenceFromProfile(profile)
	if err != nil || !reference.Valid() || !identity.Name(name) {
		return nil, "", identity.ErrInvalid
	}
	ceremonyUser := identity.CeremonyUser{ID: reference.ID(), Name: name}
	if bind {
		if _, err = s.store.User(ctx, reference.ID()); err != nil {
			return nil, "", err
		}
		for page := int64(1); ; page++ {
			items, pageErr := s.store.Passkeys(ctx, reference.ID(), page)
			if pageErr != nil {
				return nil, "", pageErr
			}
			ceremonyUser.Credentials = append(ceremonyUser.Credentials, s.credentials.RegistrationExclusions(items)...)
			if len(items) < 10 {
				break
			}
		}
	}
	options, sessionData, err := s.verifier.BeginRegistration(ceremonyUser)
	if err != nil {
		return nil, "", err
	}
	var payload identity.SessionPayload
	registration := identity.PasskeySession{User: ceremonyUser, Profile: reference, Data: sessionData, Name: name}
	if bind {
		payload = identity.PasskeyBindSession{PasskeySession: registration}
	} else {
		payload = registration
	}
	body, err := payload.Encode()
	if err != nil {
		return nil, "", err
	}
	id, err := application.OpaqueID()
	if err != nil {
		return nil, "", err
	}
	session, err := identity.NewSession(id, payload.SessionKind(), reference.ID(), body, s.now().Add(5*time.Minute))
	if err != nil {
		return nil, "", err
	}
	if err = s.store.SaveSession(ctx, session); err != nil {
		return nil, "", err
	}
	return options, id, nil
}

func (s *Service) FinishRegistration(ctx context.Context, id string, actor int64, register bool, credential []byte, locale, language string) (identity.Passkey, account.Profile, error) {
	if register {
		if len(id) > 128 || len(credential) == 0 || len(credential) > 1<<20 {
			return identity.Passkey{}, account.Profile{}, identity.ErrInvalid
		}
	} else if actor <= 0 || len(id) > 128 || len(credential) == 0 || len(credential) > 1<<20 {
		return identity.Passkey{}, account.Profile{}, identity.ErrInvalid
	}
	kind := identity.SessionKindPasskeyBind
	if register {
		kind = identity.SessionKindPasskeyRegistration
	}
	session, err := s.store.ConsumeSession(ctx, id, kind, actor, s.now())
	if err != nil {
		return identity.Passkey{}, account.Profile{}, err
	}
	var (
		data     identity.PasskeySession
		verified identity.VerifiedCredential
	)
	if register {
		decoded, decodeErr := identity.DecodePasskeySession(session.Data())
		if decodeErr != nil {
			return identity.Passkey{}, account.Profile{}, decodeErr
		}
		updated, localeErr := decoded.Profile.WithLocale(locale, language)
		if localeErr != nil {
			return identity.Passkey{}, account.Profile{}, localeErr
		}
		decoded.Profile = updated
		verified, err = s.verifier.FinishRegistration(decoded.User, decoded.Data, credential)
		data = decoded
	} else {
		decoded, decodeErr := identity.DecodePasskeyBindSession(session.Data())
		if decodeErr != nil {
			return identity.Passkey{}, account.Profile{}, decodeErr
		}
		if !decoded.Profile.Valid() {
			return identity.Passkey{}, account.Profile{}, identity.ErrInvalid
		}
		verified, err = s.verifier.FinishRegistration(decoded.User, decoded.Data, credential)
		data = decoded.PasskeySession
	}
	if err != nil {
		return identity.Passkey{}, account.Profile{}, application.ErrCredentials
	}
	profile, err := account.NewMinimalProfile(data.Profile.ID(), data.Profile.Locale(), data.Profile.Language())
	if err != nil {
		return identity.Passkey{}, account.Profile{}, err
	}
	key, err := identity.NewPasskey(verified.ID, session.OwnerID(), verified.Data, data.Name, verified.Count, s.now().UTC())
	if err != nil {
		return identity.Passkey{}, account.Profile{}, err
	}
	if err = s.credentials.ValidateNewPasskey(data.Profile, key); err != nil {
		return identity.Passkey{}, account.Profile{}, err
	}
	if err = s.store.AddPasskey(ctx, profile, key, register); err != nil {
		return identity.Passkey{}, account.Profile{}, err
	}
	s.publisher.Publish(event.CredentialAdded{
		Base:      event.NewBase(s.now().UTC()),
		AccountID: key.OwnerID(),
		PasskeyID: key.ID(),
		Bind:      !register,
	})
	return key, profile, nil
}

func (s *Service) StartLogin(ctx context.Context) ([]byte, string, error) {
	options, data, err := s.verifier.BeginLogin()
	if err != nil {
		return nil, "", err
	}
	id, err := application.OpaqueID()
	if err != nil {
		return nil, "", err
	}
	payload := identity.PasskeyLoginSession{Data: data}
	body, err := payload.Encode()
	if err != nil {
		return nil, "", err
	}
	session, err := identity.NewSession(id, payload.SessionKind(), 0, body, s.now().Add(5*time.Minute))
	if err != nil {
		return nil, "", err
	}
	if err = s.store.SaveSession(ctx, session); err != nil {
		return nil, "", err
	}
	return options, id, nil
}

func (s *Service) FinishLogin(ctx context.Context, id string, assertion []byte) (account.Profile, error) {
	if len(id) > 128 || len(assertion) == 0 || len(assertion) > 1<<20 {
		return account.Profile{}, identity.ErrInvalid
	}
	session, err := s.store.ConsumeSession(ctx, id, identity.SessionKindPasskeyLogin, 0, s.now())
	if err != nil {
		return account.Profile{}, err
	}
	stored, err := identity.DecodePasskeyLoginSession(session.Data())
	if err != nil {
		return account.Profile{}, err
	}
	var original identity.Passkey
	verified, err := s.verifier.FinishLogin(ctx, stored.Data, assertion, func(key string, handle []byte) (identity.CeremonyUser, error) {
		found, findErr := s.store.Credential(ctx, key)
		if findErr != nil {
			return identity.CeremonyUser{}, findErr
		}
		if !found.Enabled() || len(handle) != 8 || binary.BigEndian.Uint64(handle) != uint64(found.OwnerID()) {
			return identity.CeremonyUser{}, application.ErrCredentials
		}
		original = found
		return identity.CeremonyUser{ID: found.OwnerID(), Name: fmt.Sprint(found.OwnerID()), Credentials: [][]byte{found.Data()}}, nil
	})
	if err != nil || verified.ID != original.ID() {
		return account.Profile{}, application.ErrCredentials
	}
	updated, err := original.ApplyAssertion(verified, original.Version(), s.now().UTC())
	if err != nil {
		return account.Profile{}, application.ErrCredentials
	}
	if err = s.store.UpdatePasskey(ctx, updated, original.Version()); err != nil {
		return account.Profile{}, err
	}
	user, err := s.store.User(ctx, original.OwnerID())
	if err != nil {
		return account.Profile{}, err
	}
	return user.Profile(), nil
}
