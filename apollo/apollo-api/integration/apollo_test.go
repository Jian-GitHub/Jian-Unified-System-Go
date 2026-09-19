package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"jian-unified-system/apollo/apollo-rpc/apollo"

	driver "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v3"
)

var integrationConfig = flag.String("integration-config", "", "opt in: direct YAML for disposable loopback apollo_test database")

type envelope struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
}
type harness struct {
	t      *testing.T
	base   string
	client *http.Client
	db     *sql.DB
	rpc    *grpc.ClientConn
	secret string
}

func (h *harness) post(path string, body any, token string, expected int) json.RawMessage {
	h.t.Helper()
	data, _ := json.Marshal(body)
	r, _ := http.NewRequest("POST", h.base+path, bytes.NewReader(data))
	r.Header.Set("Content-Type", "application/json")
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := h.client.Do(r)
	if err != nil {
		h.t.Fatal(err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode != expected {
		h.t.Fatalf("%s status %d, expected %d: %s", path, res.StatusCode, expected, raw)
	}
	var e envelope
	if err = json.Unmarshal(raw, &e); err != nil {
		h.t.Fatalf("%s invalid envelope: %v", path, err)
	}
	if e.Code != expected {
		h.t.Fatalf("%s envelope status %d", path, e.Code)
	}
	return e.Data
}
func decode[T any](t *testing.T, raw []byte) T {
	t.Helper()
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	return out
}
func (h *harness) user(token string) int64 {
	h.t.Helper()
	p := decode[struct {
		ID string `json:"id"`
	}](h.t, h.post("/v1/account/GetUserInfo", map[string]any{}, token, 200))
	id, err := strconv.ParseInt(p.ID, 10, 64)
	if err != nil {
		h.t.Fatal(err)
	}
	return id
}
func (h *harness) authorization(provider, token string) string {
	path := "/v1/thirdParty/Continue"
	if token != "" {
		path = "/v1/thirdParty/Bind"
	}
	data := decode[struct {
		URL string `json:"url"`
	}](h.t, h.post(path, map[string]any{"provider": provider}, token, 200))
	return data.URL
}
func state(t *testing.T, raw string) string {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if u.Query().Get("code_challenge_method") != "S256" || u.Query().Get("code_challenge") == "" {
		t.Fatal("PKCE missing")
	}
	return u.Query().Get("state")
}
func (h *harness) callback(provider, state, code string, expected int) string {
	h.t.Helper()
	res, err := h.client.Get(h.base + "/v1/thirdParty/Callback/" + provider + "?" + url.Values{"state": {state}, "code": {code}}.Encode())
	if err != nil {
		h.t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != expected {
		raw, _ := io.ReadAll(res.Body)
		h.t.Fatalf("callback status %d expected %d: %s", res.StatusCode, expected, raw)
	}
	if expected != 302 {
		return ""
	}
	u, err := url.Parse(res.Header.Get("Location"))
	if err != nil {
		h.t.Fatal(err)
	}
	if res.Header.Get("Cache-Control") != "no-store" {
		h.t.Fatal("callback cache policy missing")
	}
	token := u.Query().Get("token")
	if token == "" {
		h.t.Fatal("missing callback JWT")
	}
	return token
}
func freeAddress(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := l.Addr().String()
	l.Close()
	return address
}
func readYAML(t *testing.T, path string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var c map[string]any
	if err = yaml.Unmarshal(raw, &c); err != nil {
		t.Fatal(err)
	}
	return c
}
func startProcess(t *testing.T, root, temp, name, address string, c map[string]any) {
	t.Helper()
	binary := filepath.Join(temp, name)
	command := exec.Command("go", "build", "-race", "-mod=readonly", "-o", binary, "./apollo/"+name+"")
	command.Dir = root
	if out, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build %s: %s", name, out)
	}
	raw, err := yaml.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(temp, name+".yaml")
	if err = os.WriteFile(config, raw, 0600); err != nil {
		t.Fatal(err)
	}
	log, err := os.Create(filepath.Join(temp, name+".log"))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(binary, "-f", config)
	cmd.Stdout = log
	cmd.Stderr = log
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
		}
		log.Close()
		output, err := os.ReadFile(log.Name())
		if err != nil {
			t.Error("cannot inspect service race log")
		} else if bytes.Contains(output, []byte("WARNING: DATA RACE")) {
			t.Error(name + " reported a data race")
		}
	})
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case err := <-done:
			t.Fatalf("%s exited during startup: %v (log %s)", name, err, log.Name())
		default:
		}
		conn, err := net.DialTimeout("tcp", address, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("%s startup timeout", name)
}

func TestApolloComplete(t *testing.T) {
	if *integrationConfig == "" {
		t.Skip("-integration-config not set; no database, provider or process contacted")
	}
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	testConfig := readYAML(t, *integrationConfig)
	dsn := testConfig["DB"].(map[string]any)["DataSource"].(string)
	parsed, err := driver.ParseDSN(dsn)
	if err != nil || parsed.DBName != "apollo_test" || parsed.Net != "tcp" || !strings.HasPrefix(parsed.Addr, "127.0.0.1:") {
		t.Fatal("requires disposable loopback apollo_test database")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal("invalid test DSN")
	}
	defer db.Close()
	var tables int
	if err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE()").Scan(&tables); err != nil || tables != 0 {
		t.Fatal("test database must be reachable and empty; nothing overwritten")
	}
	created := []string{}
	defer func() {
		for i := len(created) - 1; i >= 0; i-- {
			if _, err := db.Exec("DROP TABLE `" + created[i] + "`"); err != nil {
				t.Error(err)
			}
		}
	}()
	for _, file := range []string{"account.sql", "identity.sql", "outbox.sql"} {
		raw, err := os.ReadFile(filepath.Join(root, "apollo/apollo-rpc/schema", file))
		if err != nil {
			t.Fatal(err)
		}
		for _, stmt := range strings.Split(string(raw), ";") {
			if strings.TrimSpace(stmt) == "" {
				continue
			}
			if _, err = db.Exec(stmt); err != nil {
				t.Fatal(err)
			}
		}
		switch file {
		case "account.sql":
			created = append(created, "user")
		case "identity.sql":
			created = append(created, "passkey", "token", "third_party", "contact", "authentication_session")
		case "outbox.sql":
			created = append(created, "event_outbox")
		}
	}
	var providerMu sync.Mutex
	challenges := map[string]string{}
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/token") {
			r.ParseForm()
			providerMu.Lock()
			challenge := challenges[r.Form.Get("code")]
			providerMu.Unlock()
			if challenge == "" || oauth2.S256ChallengeFromVerifier(r.Form.Get("code_verifier")) != challenge {
				w.WriteHeader(401)
				fmt.Fprint(w, `{"error":"invalid_grant"}`)
				return
			}
			fmt.Fprint(w, `{"access_token":"fixture-access","token_type":"Bearer"}`)
			return
		}
		if r.Header.Get("Authorization") != "Bearer fixture-access" {
			w.WriteHeader(401)
			return
		}
		switch r.URL.Path {
		case "/github/user":
			fmt.Fprint(w, `{"id":4242,"login":"fixture","name":"Fixture"}`)
		case "/github/emails":
			fmt.Fprint(w, `[{"email":"verified@example.com","primary":true,"verified":true}]`)
		case "/google/user":
			fmt.Fprint(w, `{"sub":"4242","name":"Fixture Google","email":"verified@example.com","email_verified":true}`)
		default:
			w.WriteHeader(404)
		}
	}))
	defer provider.Close()
	authorize := func(raw string) string {
		t.Helper()
		u, _ := url.Parse(raw)
		code := u.Query().Get("state")
		providerMu.Lock()
		challenges[code] = u.Query().Get("code_challenge")
		providerMu.Unlock()
		return code
	}
	temp := t.TempDir()
	rpcAddress, apiAddress := freeAddress(t), freeAddress(t)
	rpcConfig := readYAML(t, filepath.Join(root, "apollo/apollo-rpc/etc/apollorpc.yaml"))
	rpcConfig["ListenOn"] = rpcAddress
	rpcConfig["DB"] = testConfig["DB"]
	for name, raw := range rpcConfig["OAuth"].(map[string]any) {
		c := raw.(map[string]any)
		c["AuthURL"] = provider.URL + "/" + name + "/authorize"
		c["TokenURL"] = provider.URL + "/" + name + "/token"
		c["UserInfoURL"] = provider.URL + "/" + name + "/user"
		if name == "github" {
			c["EmailsURL"] = provider.URL + "/github/emails"
		}
	}
	startProcess(t, root, temp, "apollo-rpc", rpcAddress, rpcConfig)
	apiConfig := readYAML(t, filepath.Join(root, "apollo/apollo-api/etc/apollo-api.yaml"))
	apiConfig["GeoIP"].(map[string]any)["Database"] = filepath.Join(root, "jus-core/data/GeoLite2-City.mmdb")
	_, port, _ := net.SplitHostPort(apiAddress)
	portNumber, _ := strconv.Atoi(port)
	apiConfig["Port"] = portNumber
	apiConfig["ApolloRpc"].(map[string]any)["Endpoints"] = []string{rpcAddress}
	startProcess(t, root, temp, "apollo-api", apiAddress, apiConfig)
	conn, err := grpc.NewClient(rpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	jar, _ := cookiejar.New(nil)
	h := &harness{t: t, base: "http://" + apiAddress, client: &http.Client{Jar: jar, Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}, db: db, rpc: conn, secret: apiConfig["Auth"].(map[string]any)["AccessSecret"].(string)}
	ctx := context.Background()
	security := apollo.NewSecurityClient(conn)
	third := apollo.NewThirdPartyClient(conn)
	empty := map[string]any{}
	registration := map[string]any{"email": "complete@example.com", "password": "Valid123!", "confirm_password": "Valid123!", "language": "zh"}
	first := decode[struct {
		Token string `json:"token"`
	}](t, h.post("/v1/account/registration", registration, "", 200)).Token
	firstID := h.user(first)
	profileRPC, err := apollo.NewAccountClient(conn).UserInfo(ctx, &apollo.UserInfoReq{UserId: firstID})
	if err != nil {
		t.Fatal(err)
	}
	legacy := decode[map[string]json.RawMessage](t, profileRPC.UserBytes)
	if legacy["Password"] != nil || legacy["Email"] != nil {
		t.Fatal("private account fields leaked")
	}
	notification := decode[sql.NullString](t, legacy["NotificationEmail"])
	if !notification.Valid || notification.String != "complete@example.com" || profileRPC.Profile.NotificationEmail != notification.String {
		t.Fatal("legacy and typed user projections differ")
	}
	h.post("/v1/account/registration", registration, "", 409)
	h.post("/v1/account/login", map[string]any{"email": "complete@example.com", "password": "Wrong123!", "cloudflareToken": "XXXX.DUMMY.TOKEN.XXXX"}, "", 401)
	login := decode[struct {
		Token string `json:"token"`
	}](t, h.post("/v1/account/login", map[string]any{"email": "complete@example.com", "password": "Valid123!", "cloudflareToken": "XXXX.DUMMY.TOKEN.XXXX"}, "", 200)).Token
	if h.user(login) != firstID {
		t.Fatal("password login changed owner")
	}
	h.post("/v1/account/VerifyToken", empty, first, 200)
	h.post("/v1/account/VerifyToken", empty, "", 401)
	forged, _ := jwt.NewWithClaims(jwt.SigningMethodHS384, jwt.MapClaims{"id": firstID, "exp": time.Now().Add(time.Hour).Unix()}).SignedString([]byte(h.secret))
	h.post("/v1/account/VerifyToken", empty, forged, 401)
	malformed, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"id": 1.5, "exp": time.Now().Add(time.Hour).Unix()}).SignedString([]byte(h.secret))
	h.post("/v1/account/VerifyToken", empty, malformed, 401)
	h.post("/v1/account/GetUserSecurityInfo", empty, first, 200)
	t.Log("checking Figma account-management flows and primary-contact protection")
	h.post("/v1/account/UpdateName", map[string]any{"given_name": "Jian", "middle_name": "Unified", "family_name": "System"}, first, 200)
	h.post("/v1/account/UpdateBirthday", map[string]any{"year": 2000, "month": 2, "day": 29}, first, 200)
	h.post("/v1/account/UpdateBirthday", map[string]any{"year": 2001, "month": 2, "day": 29}, first, 400)
	h.post("/v1/account/UpdateLanguage", map[string]any{"language": "en"}, first, 200)
	contact := decode[struct {
		Contact struct {
			ID      string `json:"id"`
			Value   string `json:"value"`
			Primary bool   `json:"primary"`
		} `json:"contact"`
	}](t, h.post("/v1/account/AddContact", map[string]any{"value": "secondary@example.com", "type": 1}, first, 200)).Contact
	if contact.ID == "" || contact.Value != "secondary@example.com" || contact.Primary {
		t.Fatal("invalid secondary contact projection")
	}
	securityInfo := decode[struct {
		Contacts []struct {
			ID      string `json:"id"`
			Value   string `json:"value"`
			Primary bool   `json:"primary"`
		} `json:"contacts"`
	}](t, h.post("/v1/account/GetUserSecurityInfo", empty, first, 200))
	if len(securityInfo.Contacts) != 2 || !securityInfo.Contacts[0].Primary || securityInfo.Contacts[0].Value != "complete@example.com" {
		t.Fatal("primary login contact missing or mutable")
	}
	h.post("/v1/account/RemoveContact", map[string]any{"id": strconv.FormatInt(firstID, 10)}, first, 400)
	h.post("/v1/account/RemoveContact", map[string]any{"id": contact.ID}, first, 200)
	h.post("/v1/account/ChangeNotificationEmail", map[string]any{"email": "notify@example.com"}, first, 200)
	h.post("/v1/account/RemoveNotificationEmail", empty, first, 200)
	h.post("/v1/account/ChangePassword", map[string]any{"current_password": "Valid123!", "new_password": "Changed123!", "confirm_new_password": "Mismatch123!"}, first, 400)
	h.post("/v1/account/ChangePassword", map[string]any{"current_password": "Wrong123!", "new_password": "Changed123!", "confirm_new_password": "Changed123!"}, first, 401)
	h.post("/v1/account/ChangePassword", map[string]any{"current_password": "Valid123!", "new_password": "Changed123!", "confirm_new_password": "Changed123!"}, first, 200)
	h.post("/v1/account/login", map[string]any{"email": "complete@example.com", "password": "Valid123!", "cloudflareToken": "XXXX.DUMMY.TOKEN.XXXX"}, "", 401)
	changedLogin := decode[struct {
		Token string `json:"token"`
	}](t, h.post("/v1/account/login", map[string]any{"email": "complete@example.com", "password": "Changed123!", "cloudflareToken": "XXXX.DUMMY.TOKEN.XXXX"}, "", 200)).Token
	if h.user(changedLogin) != firstID {
		t.Fatal("password update changed owner")
	}
	t.Log("checking passkey signatures, replay, origin, counter and ownership")
	type ceremony struct {
		Options string `json:"options_json"`
		Session string `json:"session_id"`
	}
	start := decode[ceremony](t, h.post("/v1/passkeys/registration/start", map[string]any{"user_name": "Passkey user"}, "", 200))
	auth := newAuthenticator(t)
	credential := auth.registration(t, start.Options)
	finish := map[string]any{"session_id": start.Session, "credential": credential, "language": "zh"}
	second := decode[struct {
		Token string `json:"token"`
	}](t, h.post("/v1/passkeys/registration/finish", finish, "", 200)).Token
	secondID := h.user(second)
	h.post("/v1/passkeys/registration/finish", finish, "", 401)
	h.post("/v1/account/security/RemovePasskey", map[string]any{"id": encoded(auth.id)}, second, 409)
	h.post("/v1/account/security/RemovePasskey", map[string]any{"id": encoded(auth.id)}, first, 404)
	assertion := func(count uint32, origin string, corrupt bool, expected int) {
		t.Helper()
		s := decode[ceremony](t, h.post("/v1/passkeys/login/start", empty, "", 200))
		body := map[string]any{"session_id": s.Session, "assertion": auth.assertion(t, s.Options, count, origin, corrupt)}
		raw := h.post("/v1/passkeys/login/finish", body, "", expected)
		if expected == 200 {
			token := decode[struct {
				Token string `json:"token"`
			}](t, raw).Token
			if h.user(token) != secondID {
				t.Fatal("passkey changed owner")
			}
			h.post("/v1/passkeys/login/finish", body, "", 401)
		}
	}
	assertion(1, "http://localhost:3000", false, 200)
	assertion(2, "http://localhost:3000", true, 401)
	assertion(2, "https://evil.invalid", false, 401)
	assertion(1, "http://localhost:3000", false, 401)
	assertion(2, "http://localhost:3000", false, 200)
	bind := decode[ceremony](t, h.post("/v1/passkeys/bind/start", map[string]any{"name": "second device"}, second, 200))
	bound := newAuthenticator(t)
	bindBody := map[string]any{"session_id": bind.Session, "credential": bound.registration(t, bind.Options), "name": "ignored override"}
	h.post("/v1/passkeys/bind/finish", bindBody, first, 401)
	h.post("/v1/passkeys/bind/finish", bindBody, second, 200)
	listed := decode[struct {
		Passkeys []struct {
			ID string `json:"id"`
		} `json:"passkeys"`
	}](t, h.post("/v1/account/security/GetTenPasskeys", map[string]any{"page": 1}, second, 200))
	if len(listed.Passkeys) != 2 {
		t.Fatal("missing bound credentials")
	}
	h.post("/v1/account/security/GetTenPasskeys", map[string]any{"page": 0}, second, 400)
	// Racing deletes must leave one login method. Requests return before assertions to avoid Fatal in a goroutine.
	var wg sync.WaitGroup
	statuses := make(chan int, 2)
	for _, id := range []string{encoded(auth.id), encoded(bound.id)} {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			raw, _ := json.Marshal(map[string]string{"id": id})
			r, _ := http.NewRequest("POST", h.base+"/v1/account/security/RemovePasskey", bytes.NewReader(raw))
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Authorization", "Bearer "+second)
			res, err := h.client.Do(r)
			if err != nil {
				statuses <- 0
				return
			}
			res.Body.Close()
			statuses <- res.StatusCode
		}(id)
	}
	wg.Wait()
	close(statuses)
	counts := map[int]int{}
	for code := range statuses {
		counts[code]++
	}
	if counts[200] != 1 || counts[409] != 1 {
		t.Fatalf("concurrent last credential protection: %v", counts)
	}
	var remaining int
	if err = db.QueryRow("SELECT COUNT(*) FROM passkey WHERE user_id=? AND is_deleted=0", secondID).Scan(&remaining); err != nil || remaining != 1 {
		t.Fatal("credential invariant broken")
	}
	// A freshly signed assertion with a deleted credential must fail.
	var removed string
	if err = db.QueryRow("SELECT credential_id FROM passkey WHERE user_id=? AND is_deleted=1", secondID).Scan(&removed); err != nil {
		t.Fatal(err)
	}
	revoked := auth
	if removed == encoded(bound.id) {
		revoked = bound
	}
	revokedStart := decode[ceremony](t, h.post("/v1/passkeys/login/start", empty, "", 200))
	h.post("/v1/passkeys/login/finish", map[string]any{"session_id": revokedStart.Session, "assertion": revoked.assertion(t, revokedStart.Options, 3, "http://localhost:3000", false)}, "", 401)
	t.Log("checking scoped grants, expiry, owner and revocation")
	tokenData := decode[struct {
		Token struct {
			ID    string `json:"id"`
			Value string `json:"value"`
		} `json:"token"`
	}](t, h.post("/v1/account/security/GenerateSubsystemToken", map[string]any{"name": "quantum", "scope": []int64{1}}, first, 200))
	grantID, _ := strconv.ParseInt(tokenData.Token.ID, 10, 64)
	claims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(tokenData.Token.Value, claims, func(*jwt.Token) (any, error) {
		return []byte(rpcConfig["SubSystem"].(map[string]any)["AccessSecret"].(string)), nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithJSONNumber())
	if err != nil || claims["scope"].(json.Number).String() != "1" || claims["id"].(json.Number).String() != strconv.FormatInt(firstID, 10) || claims["tokenId"].(json.Number).String() != tokenData.Token.ID {
		t.Fatal("invalid grant claims")
	}
	checkGrant := func(owner int64, want bool) {
		t.Helper()
		res, err := security.ValidateSubsystemToken(ctx, &apollo.ValidateSubsystemTokenReq{UserId: owner, TokenId: grantID})
		if err != nil || res.Validated != want {
			t.Fatalf("grant validation: %v %v", res, err)
		}
	}
	checkGrant(firstID, true)
	checkGrant(secondID, false)
	h.post("/v1/account/security/GenerateSubsystemToken", map[string]any{"name": "invalid", "scope": []int64{8}}, first, 400)
	h.post("/v1/account/security/GetTenSubsystemTokens", map[string]any{"page": 1}, first, 200)
	h.post("/v1/account/security/RemoveSubsystemToken", map[string]any{"id": tokenData.Token.ID}, second, 404)
	if _, err = db.Exec("UPDATE token SET expires_at=? WHERE id=?", time.Now().Add(-time.Minute), grantID); err != nil {
		t.Fatal(err)
	}
	checkGrant(firstID, false)
	h.post("/v1/account/security/RemoveSubsystemToken", map[string]any{"id": tokenData.Token.ID}, first, 200)
	checkGrant(firstID, false)
	t.Log("checking OAuth PKCE, state consumption, provider separation, binding and removal")
	authURL := h.authorization("github", "")
	code := authorize(authURL)
	// A callback URL opened in a different browser must fail without consuming RPC state.
	stranger := &http.Client{Timeout: time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	cross, err := stranger.Get(h.base + "/v1/thirdParty/Callback/github?" + url.Values{"state": {state(t, authURL)}, "code": {code}}.Encode())
	if err != nil {
		t.Fatal(err)
	}
	cross.Body.Close()
	if cross.StatusCode != 401 {
		t.Fatal("cross-browser callback accepted")
	}

	github := h.callback("github", state(t, authURL), code, 302)
	githubID := h.user(github)
	h.callback("github", state(t, authURL), code, 401)
	// Repeated provider login resolves the same account despite a newly allocated candidate ID.
	authURL = h.authorization("github", "")
	code = authorize(authURL)
	if h.user(h.callback("github", state(t, authURL), code, 302)) != githubID {
		t.Fatal("OAuth owner changed")
	}
	authURL = h.authorization("google", "")
	code = authorize(authURL)
	google := h.callback("google", state(t, authURL), code, 302)
	googleID := h.user(google)
	if googleID == githubID {
		t.Fatal("providers or equal verified email merged")
	}
	type externalList struct {
		Accounts []struct {
			ID       int64  `json:"id"`
			Provider string `json:"provider"`
		} `json:"accounts"`
	}
	external := decode[externalList](t, h.post("/v1/thirdParty/GetInfo", empty, github, 200))
	if len(external.Accounts) != 1 {
		t.Fatal("missing external identity")
	}
	h.post("/v1/thirdParty/Remove", map[string]any{"thirdPartyId": external.Accounts[0].ID}, github, 409)
	h.post("/v1/thirdParty/Remove", map[string]any{"thirdPartyId": external.Accounts[0].ID}, first, 404)
	// Already owned identities cannot be attached to another account.
	authURL = h.authorization("github", first)
	code = authorize(authURL)
	h.callback("github", state(t, authURL), code, 409)
	// Bind a passkey, then unlink OAuth and attach the released identity to the password account.
	bind = decode[ceremony](t, h.post("/v1/passkeys/bind/start", map[string]any{"name": "recovery"}, github, 200))
	recovery := newAuthenticator(t)
	h.post("/v1/passkeys/bind/finish", map[string]any{"session_id": bind.Session, "credential": recovery.registration(t, bind.Options)}, github, 200)
	h.post("/v1/thirdParty/Remove", map[string]any{"thirdPartyId": external.Accounts[0].ID}, github, 200)
	authURL = h.authorization("github", first)
	code = authorize(authURL)
	if h.user(h.callback("github", state(t, authURL), code, 302)) != firstID {
		t.Fatal("binding changed owner")
	}
	// Direct generated RPC Continue and Bind also enforce the new state/code contract.
	authURL = h.authorization("github", "")
	code = authorize(authURL)
	continued, err := third.Continue(ctx, &apollo.ThirdPartyContinueReq{Provider: "github", Code: code, State: state(t, authURL)})
	if err != nil || continued.UserId != firstID {
		t.Fatalf("Continue RPC: %v", err)
	}
	authURL = h.authorization("github", first)
	code = authorize(authURL)
	_, err = third.Bind(ctx, &apollo.ThirdPartyBindReq{Provider: "github", Code: code, State: state(t, authURL), UserId: firstID})
	if err != nil {
		t.Fatalf("Bind RPC: %v", err)
	}
	_, err = third.Continue(ctx, &apollo.ThirdPartyContinueReq{Provider: "github", Token: []byte("untrusted"), RedisDataJson: `{"id":1}`})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatal("legacy untrusted OAuth fields accepted")
	}
	authURL = h.authorization("github", "")
	h.callback("github", state(t, authURL), "invalid-code", 401)
	authURL = h.authorization("github", "")
	code = authorize(authURL)
	if _, err = db.Exec("UPDATE authentication_session SET expires_at=? WHERE id=?", time.Now().Add(-time.Minute), state(t, authURL)); err != nil {
		t.Fatal(err)
	}
	h.callback("github", state(t, authURL), code, 401)
	h.post("/v1/thirdParty/Continue", map[string]any{"provider": "unknown"}, "", 400)
	finalSecurity := decode[struct {
		Providers struct {
			Github bool `json:"github"`
		} `json:"thirdPartyAccounts"`
	}](t, h.post("/v1/account/GetUserSecurityInfo", empty, first, 200))
	if !finalSecurity.Providers.Github {
		t.Fatal("security read model missing binding")
	}
	h.post("/v1/account/ChangePassword", map[string]any{"current_password": "Changed123!", "new_password": "Final123!", "confirm_new_password": "Final123!", "sign_out_everywhere": true}, first, 200)
	h.post("/v1/account/VerifyToken", empty, first, 401)
	renewed := decode[struct {
		Token string `json:"token"`
	}](t, h.post("/v1/account/login", map[string]any{"email": "complete@example.com", "password": "Final123!", "cloudflareToken": "XXXX.DUMMY.TOKEN.XXXX"}, "", 200)).Token
	if h.user(renewed) != firstID {
		t.Fatal("session-version renewal changed owner")
	}
	h.post("/v1/account/DeleteAccount", map[string]any{"confirmation": "DELETE"}, google, 200)
	h.post("/v1/account/VerifyToken", empty, google, 401)
	var orphans int
	err = db.QueryRow("SELECT COUNT(*) FROM `user` u WHERE password='' AND NOT EXISTS(SELECT 1 FROM passkey p WHERE p.user_id=u.id AND p.is_deleted=0) AND NOT EXISTS(SELECT 1 FROM third_party p WHERE p.user_id=u.id)").Scan(&orphans)
	if err != nil || orphans != 0 {
		t.Fatal("orphaned account after failed credential operation")
	}
	t.Log("all 30 HTTP routes and all 30 RPC methods exercised through actual service processes")
}
