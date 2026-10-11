package mcp

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"custos/internal/project"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestHTTPBudgetsAndLocalIdentityScope(t *testing.T) {
	s, _, _ := fixture(t)
	token := strings.Repeat("x", 48)
	s.options.Hosts = []string{"localhost:8765"}
	s.options.Credentials = []Credential{{SHA256: project.Fingerprint([]byte(token)), Identity: "local", Expires: time.Now().Add(time.Hour), Scopes: []string{"custos:read"}}}
	handler, err := s.HTTPHandler("test")
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "http://localhost:8765/mcp", strings.NewReader(strings.Repeat("x", requestLimit+1)))
	// The scheme matches case-insensitively: reaching the body budget proves it.
	request.Header.Set("Authorization", "bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json, text/event-stream")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, request)
	if rr.Code != 413 {
		t.Fatal(rr.Code, rr.Body.String())
	}
	id, write, err := s.authenticate(context.Background(), token)
	if err != nil || id != "local" || write {
		t.Fatal(id, write, err)
	}
	ctx := context.WithValue(context.WithValue(context.Background(), identityKey{}, id), writeScopeKey{}, write)
	if _, err := s.selectWorkspace(ctx, "default", true); err == nil {
		t.Fatal("local named remote identity bypassed write scope")
	}
	s.options.Credentials = make([]Credential, 257)
	if err := s.validateAuth(); err == nil {
		t.Fatal("credential limit")
	}
	s.options.Credentials = nil
	t.Setenv("OAUTH_BUDGET_CLIENT", "client")
	t.Setenv("OAUTH_BUDGET_SECRET", "secret")
	entered := make(chan struct{}, 4)
	release := make(chan struct{})
	var wg sync.WaitGroup
	s.options.PublicURL = "https://localhost:8765/mcp"
	s.options.OAuth = &OAuth{Issuer: "https://provider.test", Introspection: "https://provider.test/introspect", Audience: s.options.PublicURL, ClientIDEnv: "OAUTH_BUDGET_CLIENT", ClientSecretEnv: "OAUTH_BUDGET_SECRET", client: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		entered <- struct{}{}
		<-release
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"active":false}`)), Header: make(http.Header)}, nil
	})}}
	handler, err = s.HTTPHandler("test")
	if err != nil {
		t.Fatal(err)
	}
	for range 4 {
		wg.Go(func() {
			r := httptest.NewRequest(http.MethodPost, "https://localhost:8765/mcp", http.NoBody)
			r.Header.Set("Authorization", "Bearer opaque")
			handler.ServeHTTP(httptest.NewRecorder(), r)
		})
	}
	for range 4 {
		<-entered
	}
	r := httptest.NewRequest(http.MethodPost, "https://localhost:8765/mcp", http.NoBody)
	r.Header.Set("Authorization", "Bearer opaque")
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, r)
	if rr.Code != 429 {
		t.Fatal(rr.Code)
	}
	close(release)
	wg.Wait()
}

func TestHTTPProtocolRevisions(t *testing.T) {
	for _, revision := range sdk.SupportedProtocolVersions() {
		t.Run(revision, func(t *testing.T) {
			s, _, _ := fixture(t)
			token := strings.Repeat("p", 48)
			s.options.Credentials = []Credential{{SHA256: project.Fingerprint([]byte(token)), Identity: "alice", Expires: time.Now().Add(time.Hour), Scopes: []string{"custos:read"}}}
			var handler http.Handler
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { handler.ServeHTTP(w, r) }))
			defer server.Close()
			s.options.Hosts = []string{strings.TrimPrefix(server.URL, "http://")}
			var err error
			handler, err = s.HTTPHandler("test")
			if err != nil {
				t.Fatal(err)
			}
			client := sdk.NewClient(&sdk.Implementation{Name: "test", Version: "1"}, nil)
			session, err := client.Connect(context.Background(), &sdk.StreamableClientTransport{Endpoint: server.URL + "/mcp", HTTPClient: &http.Client{Transport: bearerTransport{token: token}}}, &sdk.ClientSessionOptions{ProtocolVersion: revision})
			if err != nil {
				t.Fatal(err)
			}
			defer session.Close()
			tools, err := session.ListTools(context.Background(), nil)
			if err != nil || len(tools.Tools) != 9 {
				t.Fatal(tools, err)
			}
			result, err := session.CallTool(context.Background(), &sdk.CallToolParams{Name: "custos_analyze", Arguments: map[string]any{}})
			if err != nil || result.IsError {
				t.Fatal(result, err)
			}
		})
	}
}

func TestConfiguredProxyHost(t *testing.T) {
	s, _, _ := fixture(t)
	token := strings.Repeat("p", 48)
	s.options.Hosts = []string{"custos.example"}
	s.options.Credentials = []Credential{{SHA256: project.Fingerprint([]byte(token)), Identity: "alice", Expires: time.Now().Add(time.Hour), Scopes: []string{"custos:read"}}}
	handler, err := s.HTTPHandler("test")
	if err != nil {
		t.Fatal(err)
	}
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { r.Host = "custos.example"; handler.ServeHTTP(w, r) }))
	defer proxy.Close()
	client := sdk.NewClient(&sdk.Implementation{Name: "test", Version: "1"}, nil)
	session, err := client.Connect(context.Background(), &sdk.StreamableClientTransport{Endpoint: proxy.URL + "/mcp", HTTPClient: &http.Client{Transport: bearerTransport{token: token}}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	result, err := session.CallTool(context.Background(), &sdk.CallToolParams{Name: "custos_workspace", Arguments: map[string]any{}})
	if err != nil || result.IsError {
		t.Fatal(result, err)
	}
}
