package mcp

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Credential stores a SHA-256 digest, never the private bearer credential.
type Credential struct {
	SHA256   string    `json:"sha256"`
	Identity string    `json:"identity"`
	Expires  time.Time `json:"expires"`
	Scopes   []string  `json:"scopes"`
}

// OAuth delegates token issuance to an external provider. Introspection client
// credentials are supplied by environment variable names, not repository files.
type OAuth struct {
	Issuer                 string `json:"issuer"`
	Introspection          string `json:"introspection"`
	Audience               string `json:"audience"`
	ClientIDEnv            string `json:"clientIdEnv"`
	ClientSecretEnv        string `json:"clientSecretEnv"`
	clientID, clientSecret string
	client                 *http.Client
}

func httpsURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && u.Scheme == "https" && u.Host != "" && u.User == nil && u.Fragment == ""
}

func (s *Server) validateAuth() error {
	if len(s.options.Credentials) > 256 {
		return errors.New("credential capacity exceeded")
	}
	for _, c := range s.options.Credentials {
		b, err := hex.DecodeString(c.SHA256)
		if err != nil || len(b) != 32 || (c.Identity == "" || len(c.Identity) > 512) || c.Expires.IsZero() || !contains(c.Scopes, "custos:read") {
			return errors.New("invalid bearer credential configuration")
		}
	}
	if o := s.options.OAuth; o != nil {
		o.clientID = os.Getenv(o.ClientIDEnv)
		o.clientSecret = os.Getenv(o.ClientSecretEnv)
		if o.clientID == "" || o.clientSecret == "" {
			return errors.New("OAuth introspection credentials are missing")
		}
		if !httpsURL(o.Issuer) || !httpsURL(o.Introspection) || o.Audience != s.options.PublicURL || !httpsURL(s.options.PublicURL) {
			return errors.New("OAuth requires HTTPS issuer, introspection, resource URL and audience")
		}
	}
	if len(s.options.Credentials) == 0 && s.options.OAuth == nil {
		return errors.New("HTTP requires bearer credentials or OAuth")
	}
	return nil
}

func (s *Server) authenticate(ctx context.Context, token string) (string, bool, error) {
	sum := sha256.Sum256([]byte(token))
	for _, c := range s.options.Credentials {
		want, _ := hex.DecodeString(c.SHA256)
		if subtle.ConstantTimeCompare(sum[:], want) == 1 && time.Now().Before(c.Expires) {
			return c.Identity, contains(c.Scopes, "custos:write"), nil
		}
	}
	o := s.options.OAuth
	if o == nil {
		return "", false, errors.New("invalid token")
	}
	form := url.Values{"token": {token}, "resource": {o.Audience}}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.Introspection, strings.NewReader(form.Encode()))
	if err != nil {
		return "", false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(o.clientID, o.clientSecret)
	client := &http.Client{Timeout: 5 * time.Second, CheckRedirect: rejectRedirect}
	if o.client != nil {
		client = o.client
	}
	res, err := client.Do(req)
	if err != nil {
		return "", false, errors.New("token validation failed")
	}
	defer res.Body.Close()
	b, err := io.ReadAll(io.LimitReader(res.Body, 65537))
	if err != nil || len(b) > 65536 || res.StatusCode != http.StatusOK {
		return "", false, errors.New("token validation failed")
	}
	var info struct {
		Active   bool            `json:"active"`
		Sub      string          `json:"sub"`
		Scope    string          `json:"scope"`
		Exp      int64           `json:"exp"`
		Audience json.RawMessage `json:"aud"`
		Issuer   string          `json:"iss"`
	}
	if json.Unmarshal(b, &info) != nil || !info.Active || info.Sub == "" || len(info.Sub) > 256 || info.Exp <= time.Now().Unix() || (info.Issuer != "" && info.Issuer != o.Issuer) {
		return "", false, errors.New("invalid token")
	}
	var audience string
	var audiences []string
	if json.Unmarshal(info.Audience, &audience) == nil {
		audiences = []string{audience}
	} else {
		_ = json.Unmarshal(info.Audience, &audiences)
	}
	scopes := strings.Fields(info.Scope)
	if !contains(audiences, o.Audience) || !contains(scopes, "custos:read") {
		return "", false, errors.New("invalid audience or scope")
	}
	return o.Issuer + "#" + info.Sub, contains(scopes, "custos:write"), nil
}

// HTTPHandler authenticates every request before the stateless SDK handler.
// Stateless transports prevent sessions from becoming bearer capabilities.
func (s *Server) HTTPHandler(version string) (http.Handler, error) {
	if err := s.validateAuth(); err != nil {
		return nil, err
	}
	protocol := s.Protocol(version)
	transport := sdk.NewStreamableHTTPHandler(func(*http.Request) *sdk.Server { return protocol }, &sdk.StreamableHTTPOptions{
		Stateless: true, JSONResponse: true, MaxRequestBodyBytes: requestLimit,
		// Exact host allowlisting below replaces the localhost heuristic and also
		// accepts the configured public Host forwarded by a TLS reverse proxy.
		DisableLocalhostProtection: true,
	})
	slots := make(chan struct{}, 4)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !contains(s.options.Hosts, r.Host) {
			http.Error(w, "host denied", http.StatusForbidden)
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" && !contains(s.options.Origins, origin) {
			http.Error(w, "origin denied", http.StatusForbidden)
			return
		}
		if r.URL.Path == "/.well-known/oauth-protected-resource" || r.URL.Path == "/.well-known/oauth-protected-resource/mcp" {
			if s.options.OAuth == nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(Output{"resource": s.options.PublicURL, "authorization_servers": []string{s.options.OAuth.Issuer}, "scopes_supported": []string{"custos:read", "custos:write"}, "bearer_methods_supported": []string{"header"}})
			return
		}
		if r.URL.Path != "/mcp" {
			http.NotFound(w, r)
			return
		}
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || token == "" || len(token) > 8192 || strings.ContainsAny(token, " \t\r\n") {
			unauthorized(w, s.options.PublicURL)
			return
		}
		select {
		case slots <- struct{}{}:
			defer func() { <-slots }()
		default:
			http.Error(w, "server busy", http.StatusTooManyRequests)
			return
		}
		id, write, err := s.authenticate(r.Context(), token)
		if err != nil {
			unauthorized(w, s.options.PublicURL)
			return
		}
		ctx := context.WithValue(r.Context(), identityKey{}, id)
		ctx = context.WithValue(ctx, writeScopeKey{}, write)
		ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
		transport.ServeHTTP(w, r.WithContext(ctx))
	}), nil
}

type writeScopeKey struct{}

func unauthorized(w http.ResponseWriter, resource string) {
	challenge := "Bearer"
	if httpsURL(resource) {
		u, _ := url.Parse(resource)
		u.Path = "/.well-known/oauth-protected-resource"
		u.RawQuery = ""
		challenge += " resource_metadata=\"" + u.String() + "\""
	}
	w.Header().Set("WWW-Authenticate", challenge)
	http.Error(w, "unauthorized", http.StatusUnauthorized)
}

func rejectRedirect(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
