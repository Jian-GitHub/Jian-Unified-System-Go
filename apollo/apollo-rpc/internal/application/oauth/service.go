package oauth

import (
	"context"
	"time"

	"jian-unified-system/apollo/apollo-rpc/internal/application"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/account"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/event"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
)

// Service owns OAuth authorization, callback resolution and identity bindings.
type Service struct {
	store       identity.Store
	providers   map[string]identity.OAuth
	credentials identity.CredentialService
	publisher   event.Publisher
	now         func() time.Time
}

func NewService(store identity.Store, providers map[string]identity.OAuth, publisher event.Publisher) *Service {
	if publisher == nil {
		publisher = event.NoopPublisher{}
	}
	return &Service{
		store:       store,
		providers:   providers,
		credentials: identity.NewCredentialService(),
		publisher:   publisher,
		now:         time.Now,
	}
}

func (s *Service) StartAuthorization(ctx context.Context, provider string, profile account.Profile, bind bool) (string, error) {
	implementation, ok := s.providers[provider]
	if !ok || !identity.Provider(provider) {
		return "", identity.ErrInvalid
	}
	reference, err := referenceFromProfile(profile)
	if err != nil || !reference.Valid() {
		return "", identity.ErrInvalid
	}
	if bind {
		if _, err = s.store.User(ctx, reference.ID()); err != nil {
			return "", err
		}
	}
	state, err := application.OpaqueID()
	if err != nil {
		return "", err
	}
	verifier, err := application.OpaqueID()
	if err != nil {
		return "", err
	}
	payload := identity.AuthorizationSession{Provider: provider, Verifier: verifier, Profile: reference, Bind: bind}
	body, err := payload.Encode()
	if err != nil {
		return "", err
	}
	session, err := identity.NewSession(state, identity.OAuthSessionKind(provider), reference.ID(), body, s.now().Add(5*time.Minute))
	if err != nil {
		return "", err
	}
	if err = s.store.SaveSession(ctx, session); err != nil {
		return "", err
	}
	return implementation.AuthorizationURL(state, verifier), nil
}

// CompleteAuthorization consumes state before making provider I/O.
func (s *Service) CompleteAuthorization(ctx context.Context, provider, state, code, mode string, actor int64) (int64, error) {
	implementation, ok := s.providers[provider]
	if !ok || len(state) > 128 || state == "" {
		return 0, identity.ErrInvalid
	}
	if mode == "bind" && actor <= 0 {
		return 0, application.ErrCredentials
	}
	session, err := s.store.ConsumeSession(ctx, state, identity.OAuthSessionKind(provider), actor, s.now())
	if err != nil {
		return 0, err
	}
	data, err := identity.DecodeAuthorizationSession(session.Data())
	if err != nil {
		return 0, err
	}
	if data.Provider != provider || mode == "login" && data.Bind || mode == "bind" && !data.Bind || code == "" || len(code) > 4096 {
		return 0, application.ErrCredentials
	}
	external, err := implementation.Exchange(ctx, code, data.Verifier)
	if err != nil {
		return 0, err
	}
	if err = external.Validate(); err != nil || external.Provider() != provider {
		return 0, application.ErrCredentials
	}
	profile, err := account.NewMinimalProfile(data.Profile.ID(), data.Profile.Locale(), data.Profile.Language())
	if err != nil {
		return 0, err
	}
	owner, err := s.store.ResolveIdentity(ctx, external, profile, data.Bind)
	if err != nil {
		return owner, err
	}
	s.publisher.Publish(event.IdentityResolved{
		Base:      event.NewBase(s.now().UTC()),
		AccountID: owner,
		Provider:  provider,
		Bind:      data.Bind,
	})
	return owner, nil
}

func (s *Service) ExternalAccounts(ctx context.Context, id int64) ([]identity.ExternalAccount, error) {
	if id <= 0 {
		return nil, identity.ErrInvalid
	}
	return s.store.ExternalAccounts(ctx, id)
}

func (s *Service) RemoveIdentity(ctx context.Context, id, key int64) error {
	if id <= 0 || key <= 0 {
		return identity.ErrInvalid
	}
	accounts, err := s.store.ExternalAccounts(ctx, id)
	if err != nil {
		return err
	}
	found := false
	for _, externalAccount := range accounts {
		if externalAccount.ID() == key {
			found = true
			break
		}
	}
	if !found {
		return application.ErrNotFound
	}
	lock, err := s.store.LockAccount(ctx, id)
	if err != nil {
		return err
	}
	defer lock.Rollback()
	inventory, err := s.store.CountCredentials(ctx, lock)
	if err != nil {
		return err
	}
	if err = s.credentials.CanRemoveIdentity(inventory); err != nil {
		return err
	}
	if err = s.store.DeleteIdentity(ctx, lock, key); err != nil {
		return err
	}
	return lock.Commit()
}
