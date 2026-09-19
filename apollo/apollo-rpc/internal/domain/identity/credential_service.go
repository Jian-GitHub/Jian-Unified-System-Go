package identity

// CredentialService coordinates rules that involve a credential aggregate and
// an account-wide inventory snapshot.
type CredentialService struct{}

func NewCredentialService() CredentialService { return CredentialService{} }

// ValidateNewPasskey ensures the anti-corruption reference and credential
// aggregate identify the same account before persistence provisions or binds it.
func (CredentialService) ValidateNewPasskey(reference AccountReference, credential Passkey) error {
	if !reference.Valid() || reference.ID() != credential.OwnerID() {
		return ErrInvalid
	}
	return nil
}

func (CredentialService) CanRemovePasskey(credential Passkey, inventory CredentialInventory) error {
	return credential.IsRemovable(inventory)
}

func (CredentialService) CanRemoveIdentity(inventory CredentialInventory) error {
	return CanRemoveCredential(inventory.Count())
}

func (CredentialService) RegistrationExclusions(credentials []Passkey) [][]byte {
	return EnabledCredentials(credentials)
}
