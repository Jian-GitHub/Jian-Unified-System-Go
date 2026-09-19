// Package grant orchestrates subsystem access-grant use cases.
package grant

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"jian-unified-system/apollo/apollo-rpc/internal/application"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/event"
	"jian-unified-system/apollo/apollo-rpc/internal/domain/identity"
)

// Service owns subsystem grant issuance, validation, listing and revocation.
type Service struct {
	store     identity.Store
	signer    identity.GrantSigner
	ttl       time.Duration
	publisher event.Publisher
	now       func() time.Time
}

func NewService(store identity.Store, signer identity.GrantSigner, ttl time.Duration, publisher event.Publisher) *Service {
	if publisher == nil {
		publisher = event.NoopPublisher{}
	}
	return &Service{store: store, signer: signer, ttl: ttl, publisher: publisher, now: time.Now}
}

func (s *Service) List(ctx context.Context, id, page int64) ([]identity.Grant, error) {
	if id <= 0 || !identity.Page(page) {
		return nil, identity.ErrInvalid
	}
	return s.store.Grants(ctx, id, page)
}

func (s *Service) Remove(ctx context.Context, id, grantID int64) error {
	if id <= 0 || grantID <= 0 {
		return identity.ErrInvalid
	}
	lock, err := s.store.LockAccount(ctx, id)
	if err != nil {
		return err
	}
	defer lock.Rollback()
	if err = s.store.DeleteGrant(ctx, lock, grantID); err != nil {
		return err
	}
	if err = lock.Commit(); err != nil {
		return err
	}
	s.publisher.Publish(event.GrantRevoked{
		Base:      event.NewBase(s.now().UTC()),
		AccountID: id,
		GrantID:   grantID,
	})
	return nil
}

func (s *Service) Create(ctx context.Context, id int64, name string, scopes []int64) (identity.Grant, error) {
	if strings.TrimSpace(name) == "" {
		generated, err := uuid.NewRandom()
		if err != nil {
			return identity.Grant{}, err
		}
		name = strings.ToUpper(generated.String()[:5])
	}
	key, err := application.RandomID()
	if err != nil {
		return identity.Grant{}, err
	}
	created, err := identity.NewGrant(key, id, name, scopes, s.now().UTC(), s.ttl)
	if err != nil {
		return identity.Grant{}, err
	}
	value, err := s.signer.Sign(created)
	if err != nil {
		return identity.Grant{}, err
	}
	created, err = created.WithValue(value)
	if err != nil {
		return identity.Grant{}, err
	}
	if err = s.store.CreateGrant(ctx, created); err != nil {
		return identity.Grant{}, err
	}
	s.publisher.Publish(event.GrantIssued{
		Base:      event.NewBase(s.now().UTC()),
		AccountID: created.OwnerID(),
		GrantID:   created.ID(),
		Name:      created.Name(),
	})
	return created, nil
}

func (s *Service) Validate(ctx context.Context, id, key int64) (bool, error) {
	if id <= 0 || key <= 0 {
		return false, identity.ErrInvalid
	}
	found, err := s.store.Grant(ctx, id, key)
	if errors.Is(err, application.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return found.Active(id, s.now()), nil
}
