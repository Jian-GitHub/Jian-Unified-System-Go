package identity

import (
	"errors"
	"testing"
	"time"
)

func TestCredentialServiceCoordinatesAccountInventory(t *testing.T) {
	service := NewCredentialService()
	reference, err := NewAccountReference(7, "CN", "zh")
	if err != nil {
		t.Fatal(err)
	}
	credential, err := NewPasskey("key", 7, []byte("credential"), "phone", 0, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err = service.ValidateNewPasskey(reference, credential); err != nil {
		t.Fatal(err)
	}
	other, err := NewAccountReference(8, "CN", "zh")
	if err != nil {
		t.Fatal(err)
	}
	if !errors.Is(service.ValidateNewPasskey(other, credential), ErrInvalid) {
		t.Fatal("cross-account credential accepted")
	}

	last, err := NewCredentialInventory(false, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !errors.Is(service.CanRemovePasskey(credential, last), ErrLastCredential) {
		t.Fatal("last passkey removal accepted")
	}
	multiple, err := NewCredentialInventory(true, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err = service.CanRemoveIdentity(multiple); err != nil {
		t.Fatal(err)
	}

	disabled, err := RestorePasskey("disabled", 7, []byte("d"), "old", 0, 0, time.Now(), false)
	if err != nil {
		t.Fatal(err)
	}
	exclusions := service.RegistrationExclusions([]Passkey{credential, disabled})
	if len(exclusions) != 1 || string(exclusions[0]) != "credential" {
		t.Fatal("registration exclusions included disabled credentials")
	}
}
