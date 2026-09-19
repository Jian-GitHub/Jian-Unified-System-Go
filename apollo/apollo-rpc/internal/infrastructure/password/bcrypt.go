package password

import "golang.org/x/crypto/bcrypt"

type Bcrypt struct{ dummy []byte }

func New() (*Bcrypt, error) {
	dummy, err := bcrypt.GenerateFromPassword([]byte("unused-comparison-password"), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &Bcrypt{dummy: dummy}, nil
}

func (b *Bcrypt) Hash(value string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(value), bcrypt.DefaultCost)
	return string(hash), err
}

func (b *Bcrypt) Verify(value, hash string) bool {
	if hash == "" {
		_ = bcrypt.CompareHashAndPassword(b.dummy, []byte(value))
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(value)) == nil
}
