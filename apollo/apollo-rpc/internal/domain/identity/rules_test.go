package identity

import (
	"errors"
	"testing"
	"time"
)

func TestCredentialAndGrantRules(t *testing.T) {
	for _, count := range []int64{0, 1} {
		if !errors.Is(CanRemoveCredential(count), ErrLastCredential) {
			t.Fatal("last credential can be removed")
		}
	}
	if CanRemoveCredential(2) != nil {
		t.Fatal("second credential cannot be removed")
	}
	for _, pair := range [][2]uint32{{0, 0}, {0, 1}, {10, 11}} {
		if !CounterValid(pair[0], pair[1]) {
			t.Fatal("valid counter rejected")
		}
	}
	for _, pair := range [][2]uint32{{1, 0}, {1, 1}, {10, 9}} {
		if CounterValid(pair[0], pair[1]) {
			t.Fatal("counter regression accepted")
		}
	}
	now := time.Now()
	g, err := NewGrant(1, 2, "quantum", []int64{1, 1}, now, time.Hour)
	if err != nil || g.Scope() != 1 {
		t.Fatal("scope union failed")
	}
	if !g.Active(2, now) || g.Active(3, now) || g.Active(2, now.Add(time.Hour)) {
		t.Fatal("owner or expiry rule failed")
	}
	revoked := g.Revoke()
	if revoked.Active(2, now) || !g.Active(2, now) {
		t.Fatal("grant revocation is not immutable")
	}
	disabled := g.Disable()
	if disabled.Active(2, now) || !g.Active(2, now) {
		t.Fatal("grant disable is not immutable")
	}
	for _, scopes := range [][]int64{nil, {0}, {8}, {-1}, {1, 8}, {3}, {7}} {
		if _, err = NewGrant(1, 2, "quantum", scopes, now, time.Hour); !errors.Is(err, ErrInvalid) {
			t.Fatal("unknown scope accepted")
		}
	}
	for _, page := range []int64{-1, 0, 100001} {
		if Page(page) {
			t.Fatal("unbounded page accepted")
		}
	}
}

func TestSessionPayloadKinds(t *testing.T) {
	cases := []struct {
		name string
		want string
		got  SessionPayload
	}{
		{"register", SessionKindPasskeyRegistration, PasskeySession{}},
		{"bind", SessionKindPasskeyBind, PasskeyBindSession{}},
		{"login", SessionKindPasskeyLogin, PasskeyLoginSession{}},
	}
	for _, c := range cases {
		if c.got.SessionKind() != c.want {
			t.Fatalf("%s kind = %q, want %q", c.name, c.got.SessionKind(), c.want)
		}
	}
	oauthEmpty := AuthorizationSession{}
	if oauthEmpty.SessionKind() != "oauth" {
		t.Fatal("oauth prefix changed")
	}
	// AuthorizationSession.SessionKind is the prefix; application composes the
	// provider-aware kind with OAuthSessionKind(provider).
	if OAuthSessionKind("github") != "oauth.github" || OAuthSessionKind("google") != "oauth.google" {
		t.Fatal("OAuthSessionKind provider composition broken")
	}
}

func TestSessionProfileAndProvisioning(t *testing.T) {
	empty := SessionProfile{}
	if empty.Valid() {
		t.Fatal("empty profile accepted")
	}
	filled, err := NewSessionProfile(1, "zh-CN", "zh")
	if err != nil {
		t.Fatal(err)
	}
	if !filled.Valid() {
		t.Fatal("filled profile rejected")
	}
	oversize, oerr := NewSessionProfile(1, bytesRepeat(33), "zh")
	if oerr == nil || oversize.Valid() {
		t.Fatal("oversize locale accepted")
	}
	if _, err := NewSessionProfile(0, "zh-CN", "zh"); err == nil {
		t.Fatal("zero id accepted")
	}
	if _, err := NewSessionProfile(1, "", "zh"); err == nil {
		t.Fatal("empty locale accepted")
	}
	noNotify := AccountProvisioning{}
	if noNotify.HasNotification() {
		t.Fatal("empty provisioning reports notification")
	}
	encrypted := NewAccountProvisioning(SessionProfile{}, "a@b.c", true)
	if !encrypted.HasNotification() {
		t.Fatal("encrypted email ignored")
	}
	plaintext := NewAccountProvisioning(SessionProfile{}, "a@b.c", false)
	if plaintext.HasNotification() {
		t.Fatal("plaintext email treated as notification")
	}
	replaced, werr := filled.WithLocale("en-US", "en")
	if werr != nil {
		t.Fatal(werr)
	}
	if replaced.Locale() != "en-US" || filled.Locale() != "zh-CN" {
		t.Fatal("WithLocale leaked into the original")
	}
	if _, werr := filled.WithLocale("", "en"); werr == nil {
		t.Fatal("WithLocale accepted empty locale")
	}
}

func TestSessionEncodeDecodeRoundTrip(t *testing.T) {
	profile, perr := NewSessionProfile(7, "zh-CN", "zh")
	if perr != nil {
		t.Fatal(perr)
	}
	reg := PasskeySession{User: CeremonyUser{ID: 7, Name: "alice"}, Profile: profile, Data: []byte("challenge"), Name: "MacBook"}
	body, err := reg.Encode()
	if err != nil {
		t.Fatal(err)
	}
	back, err := DecodePasskeySession(body)
	if err != nil {
		t.Fatal(err)
	}
	if back.Name != reg.Name || back.Profile.Locale() != reg.Profile.Locale() || string(back.Data) != "challenge" {
		t.Fatal("register session round-trip lost data")
	}

	bind := PasskeyBindSession{PasskeySession: reg}
	if bind.SessionKind() != SessionKindPasskeyBind {
		t.Fatal("bind session kind mismatch")
	}
	bindBody, err := bind.Encode()
	if err != nil {
		t.Fatal(err)
	}
	bindBack, err := DecodePasskeyBindSession(bindBody)
	if err != nil {
		t.Fatal(err)
	}
	if bindBack.PasskeySession.Name != reg.Name {
		t.Fatal("bind round-trip lost payload")
	}

	login := PasskeyLoginSession{Data: []byte("login-challenge")}
	loginBody, err := login.Encode()
	if err != nil {
		t.Fatal(err)
	}
	loginBack, err := DecodePasskeyLoginSession(loginBody)
	if err != nil {
		t.Fatal(err)
	}
	if string(loginBack.Data) != "login-challenge" {
		t.Fatal("login round-trip lost data")
	}

	authProfile, perr := NewSessionProfile(1, "en-US", "en")
	if perr != nil {
		t.Fatal(perr)
	}
	auth := AuthorizationSession{Provider: "github", Verifier: "v", Profile: authProfile, Bind: true}
	authBody, err := auth.Encode()
	if err != nil {
		t.Fatal(err)
	}
	authBack, err := DecodeAuthorizationSession(authBody)
	if err != nil {
		t.Fatal(err)
	}
	if authBack.Provider != "github" || authBack.Verifier != "v" || !authBack.Bind {
		t.Fatal("auth round-trip lost data")
	}
}

func bytesRepeat(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'x'
	}
	return string(b)
}

func TestEnabledCredentials(t *testing.T) {
	now := time.Now().UTC()
	first, err := RestorePasskey("a", 1, []byte("a-data"), "first", 0, 0, now, true)
	if err != nil {
		t.Fatal(err)
	}
	disabled, err := RestorePasskey("b", 1, []byte("b-data"), "second", 0, 0, now, false)
	if err != nil {
		t.Fatal(err)
	}
	third, err := RestorePasskey("c", 1, []byte("c-data"), "third", 0, 0, now, true)
	if err != nil {
		t.Fatal(err)
	}
	got := EnabledCredentials([]Passkey{first, disabled, third})
	if len(got) != 2 || string(got[0]) != "a-data" || string(got[1]) != "c-data" {
		t.Fatalf("EnabledCredentials selection wrong: %#v", got)
	}
	got[0][0] = 'x'
	if string(first.Data()) != "a-data" {
		t.Fatal("credential bytes escaped aggregate ownership")
	}
	if EnabledCredentials(nil) == nil {
		t.Fatal("nil input should yield non-nil empty slice")
	}
}

func TestEnsureRules(t *testing.T) {
	if EnsureCounterAcceptable(0, 0) != nil {
		t.Fatal("zero counter should be acceptable")
	}
	if EnsureCounterAcceptable(0, 1) != nil {
		t.Fatal("forward counter should be acceptable")
	}
	if !errors.Is(EnsureCounterAcceptable(1, 0), ErrInvalid) {
		t.Fatal("counter regression must be rejected")
	}
	if EnsureVersionMatch(3, 3) != nil {
		t.Fatal("matching version should pass")
	}
	if !errors.Is(EnsureVersionMatch(3, 4), ErrInvalid) {
		t.Fatal("stale version must be rejected")
	}
}

func TestPasskeyApplyAssertion(t *testing.T) {
	now := time.Now().UTC()
	original, err := RestorePasskey("cred", 1, []byte("v1"), "device", 5, 7, now.Add(-time.Hour), true)
	if err != nil {
		t.Fatal(err)
	}
	verifiedData := []byte("v2")
	updated, err := original.ApplyAssertion(VerifiedCredential{ID: "cred", Data: verifiedData, Count: 6}, 7, now)
	if err != nil {
		t.Fatal(err)
	}
	verifiedData[0] = 'x'
	if string(updated.Data()) != "v2" || updated.SignCount() != 6 || updated.Version() != 8 {
		t.Fatal("assertion did not advance the aggregate")
	}
	if original.SignCount() != 5 || original.Version() != 7 || string(original.Data()) != "v1" {
		t.Fatal("original aggregate mutated")
	}
	if _, err := original.ApplyAssertion(VerifiedCredential{ID: "cred", Data: []byte("v3"), Count: 4}, 7, now); err == nil {
		t.Fatal("counter regression accepted")
	}
	if _, err := original.ApplyAssertion(VerifiedCredential{ID: "cred", Data: []byte("v3"), Count: 6}, 6, now); err == nil {
		t.Fatal("stale version accepted")
	}
	if _, err := original.ApplyAssertion(VerifiedCredential{ID: "other", Data: []byte("v3"), Count: 6}, 7, now); err == nil {
		t.Fatal("different credential id accepted")
	}
	if _, err := original.ApplyAssertion(VerifiedCredential{ID: "cred", Data: []byte("v3"), Count: 6}, 7, time.Time{}); err == nil {
		t.Fatal("zero timestamp accepted")
	}
}

func TestCredentialInventoryAndPasskeyRemovability(t *testing.T) {
	if !(CredentialInventory{}).Empty() {
		t.Fatal("empty inventory should be zero")
	}
	passwordOnly, err := NewCredentialInventory(true, 0, 0)
	if err != nil || passwordOnly.Count() != 1 || !passwordOnly.HasOnlyPassword() {
		t.Fatal("password-only inventory must be 1")
	}
	twoPasskeys, err := NewCredentialInventory(false, 2, 0)
	if err != nil || twoPasskeys.Count() != 2 {
		t.Fatal("two passkeys must be 2")
	}
	mixed, err := NewCredentialInventory(true, 1, 1)
	if err != nil || mixed.Count() != 3 {
		t.Fatal("mixed inventory must be 3")
	}
	if _, err := NewCredentialInventory(false, -1, 0); err == nil {
		t.Fatal("negative count accepted")
	}
	single, err := NewPasskey("single", 1, []byte("credential"), "device", 0, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	onlyPasskey, _ := NewCredentialInventory(false, 1, 0)
	if !errors.Is(single.IsRemovable(onlyPasskey), ErrLastCredential) {
		t.Fatal("only passkey must block removal")
	}
	passwordAndPasskey, _ := NewCredentialInventory(true, 1, 0)
	if single.IsRemovable(passwordAndPasskey) != nil {
		t.Fatal("password alongside passkey must permit removal")
	}
}
