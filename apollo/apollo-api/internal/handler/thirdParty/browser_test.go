package thirdParty

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCallbackRequiresInitiatingBrowser(t *testing.T) {
	state := strings.Repeat("a", 43)
	r := httptest.NewRequest("POST", "https://localhost/v1/thirdParty/Continue", nil)
	w := httptest.NewRecorder()
	if err := beginBrowser(w, r, "https://provider.invalid/authorize?state="+state, "secret"); err != nil {
		t.Fatal(err)
	}
	cookie := w.Result().Cookies()[0]
	if !cookie.Secure || !cookie.HttpOnly || cookie.MaxAge != 300 || cookie.Path != "/" {
		t.Fatal("invalid browser cookie policy")
	}
	callback := httptest.NewRequest("GET", "https://localhost/v1/thirdParty/Callback/github?state="+state, nil)
	if consumeBrowser(httptest.NewRecorder(), callback, state, "secret") == nil {
		t.Fatal("cross-browser callback accepted")
	}
	callback.AddCookie(cookie)
	if err := consumeBrowser(httptest.NewRecorder(), callback, state, "secret"); err != nil {
		t.Fatal(err)
	}
	if consumeBrowser(httptest.NewRecorder(), callback, state, "wrong-key") == nil {
		t.Fatal("forged browser proof accepted")
	}
	if consumeBrowser(httptest.NewRecorder(), callback, strings.Repeat("b", 43), "secret") == nil {
		t.Fatal("different ceremony accepted")
	}
}
