package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestOAuthIntrospection(t *testing.T) {
	s, _, _ := fixture(t)
	t.Setenv("TEST_OAUTH_CLIENT", "client")
	t.Setenv("TEST_OAUTH_SECRET", "secret")
	response := `{"active":true,"sub":"subject","scope":"custos:read custos:write","exp":` + time.Now().Add(time.Hour).Format("ignored") + `}`
	status := 200
	provider := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "client" || pass != "secret" {
			t.Error("missing provider authentication")
		}
		if err := r.ParseForm(); err != nil {
			t.Error(err)
		}
		if r.Form.Get("token") != "opaque" || r.Form.Get("resource") != "https://resource.test/mcp" {
			t.Error("invalid introspection request")
		}
		if status == 302 {
			w.Header().Set("Location", "http://must-not-contact.test/introspect")
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(response))
	}))
	defer provider.Close()
	s.options.PublicURL = "https://resource.test/mcp"
	s.options.Hosts = []string{"resource.test"}
	o := &OAuth{Issuer: provider.URL, Introspection: provider.URL, Audience: s.options.PublicURL, ClientIDEnv: "TEST_OAUTH_CLIENT", ClientSecretEnv: "TEST_OAUTH_SECRET", client: provider.Client()}
	client := *provider.Client()
	client.CheckRedirect = rejectRedirect
	o.client = &client
	s.options.OAuth = o
	if err := s.validateAuth(); err != nil {
		t.Fatal(err)
	}
	active := map[string]any{"active": true, "sub": "subject", "scope": "custos:read custos:write", "exp": time.Now().Add(time.Hour).Unix(), "aud": s.options.PublicURL, "iss": provider.URL}
	set := func(info map[string]any) {
		b, err := json.Marshal(info)
		if err != nil {
			t.Fatal(err)
		}
		response = string(b)
	}
	set(active)
	id, write, err := s.authenticate(context.Background(), "opaque")
	if err != nil || !write || id != provider.URL+"#subject" {
		t.Fatal(id, write, err)
	}
	for _, tc := range []struct {
		key   string
		value any
	}{{"active", false}, {"sub", ""}, {"sub", strings.Repeat("a", 257)}, {"exp", time.Now().Add(-time.Minute).Unix()}, {"aud", "https://wrong.test"}, {"aud", []string{"wrong"}}, {"scope", "custos:write"}, {"iss", "https://wrong.test"}} {
		t.Run(tc.key+"-"+strings.TrimSpace(tc.key), func(t *testing.T) {
			bad := map[string]any{}
			for k, v := range active {
				bad[k] = v
			}
			bad[tc.key] = tc.value
			set(bad)
			if _, _, err := s.authenticate(context.Background(), "opaque"); err == nil {
				t.Fatal("invalid introspection accepted", tc)
			}
		})
	}
	active["aud"] = []string{s.options.PublicURL}
	active["scope"] = "custos:read"
	set(active)
	_, write, err = s.authenticate(context.Background(), "opaque")
	if err != nil || write {
		t.Fatal(write, err)
	}
	for _, body := range []string{"not json", strings.Repeat("x", 65537)} {
		response = body
		if _, _, err := s.authenticate(context.Background(), "opaque"); err == nil {
			t.Fatal("bad body accepted")
		}
	}
	status = 302
	if _, _, err := s.authenticate(context.Background(), "opaque"); err == nil {
		t.Fatal("redirect accepted")
	}
	status = 200
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := s.authenticate(cancelled, "opaque"); err == nil {
		t.Fatal("cancelled introspection accepted")
	}
	handler, err := s.HTTPHandler("test")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/.well-known/oauth-protected-resource", "/.well-known/oauth-protected-resource/mcp"} {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "https://resource.test"+path, http.NoBody))
		if rr.Code != 200 || !strings.Contains(rr.Body.String(), s.options.PublicURL) {
			t.Fatal(rr.Code, rr.Body.String())
		}
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "https://resource.test/mcp", http.NoBody))
	if rr.Code != 401 || !strings.Contains(rr.Header().Get("WWW-Authenticate"), "resource_metadata=") {
		t.Fatal(rr.Code, rr.Header())
	}
	o.Introspection = "https://bad\nurl"
	if _, _, err := s.authenticate(context.Background(), "opaque"); err == nil {
		t.Fatal("invalid provider URL")
	}
}

func TestAuthConfiguration(t *testing.T) {
	s, _, _ := fixture(t)
	if _, err := s.HTTPHandler("test"); err == nil {
		t.Fatal("anonymous HTTP permitted")
	}
	s.options.Credentials = []Credential{{SHA256: "invalid"}}
	if err := s.validateAuth(); err == nil {
		t.Fatal("bad credential permitted")
	}
	s.options.Credentials = nil
	s.options.OAuth = &OAuth{Issuer: "http://provider.test", Introspection: "http://provider.test", Audience: "resource", ClientIDEnv: "MISSING", ClientSecretEnv: "MISSING"}
	if err := s.validateAuth(); err == nil {
		t.Fatal("missing provider credentials")
	}
	t.Setenv("MISSING", "secret")
	if err := s.validateAuth(); err == nil {
		t.Fatal("insecure provider")
	}
	for _, raw := range []string{"http://host", "https://u:p@host", "https://host/#fragment", "garbage", ""} {
		if httpsURL(raw) {
			t.Fatal(raw)
		}
	}
}
