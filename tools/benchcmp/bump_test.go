//go:build unix

package main

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

func TestLatest(t *testing.T) {
	if got := latest([]string{"1.9.0", "v1.10.0", "1.10.0-RC1", "2.0.0-beta", "dev-main", "1.2"}); got != "v1.10.0" {
		t.Errorf("latest = %q", got)
	}
	if got := latest([]string{"1.0.0-RC1"}); got != "" {
		t.Errorf("latest of pre-releases = %q", got)
	}
}

// packagist serves p2 metadata for the given packages; others are 404, and
// "bad/json" is not JSON.
func packagist(t *testing.T, versions map[string][]string) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/p2/bad/json.json" {
			_, _ = w.Write([]byte("{"))
			return
		}
		for pkg, vs := range versions {
			if r.URL.Path == "/p2/"+pkg+".json" {
				body := `{"packages":{"` + pkg + `":[`
				for i, v := range vs {
					if i > 0 {
						body += ","
					}
					body += `{"version":"` + v + `"}`
				}
				_, _ = w.Write([]byte(body + "]}}"))
				return
			}
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestBump(t *testing.T) {
	repo := gitRepo(t, false, "1.0.0", "1.1.0", "2.0.0-RC1")
	base := packagist(t, map[string][]string{"a/a": {"2.0.0", "1.0.0"}, "b/b": {"3.0.0"}, "c/c": {"dev-main"}})
	c := &Config{
		Corpora:  []Corpus{{Name: "x", Repo: repo, Tag: "1.0.0"}},
		Packages: map[string]string{"a/a": "1.0.0", "b/b": "3.0.0"},
	}
	changes, err := bump(c, base)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"x 1.0.0 → 1.1.0", "a/a 1.0.0 → 2.0.0"}
	if !slices.Equal(changes, want) || c.Corpora[0].Tag != "1.1.0" || c.Packages["a/a"] != "2.0.0" {
		t.Errorf("bump = %q, %+v", changes, c)
	}

	for name, c := range map[string]*Config{
		"bad repo":      {Corpora: []Corpus{{Repo: "file:///does/not/exist"}}},
		"no stable tag": {Corpora: []Corpus{{Repo: gitRepo(t, false, "1.0.0-RC1")}}},
		"no stable":     {Packages: map[string]string{"c/c": "1"}},
		"unknown":       {Packages: map[string]string{"d/d": "1"}},
		"bad metadata":  {Packages: map[string]string{"bad/json": "1"}},
	} {
		if _, err := bump(c, base); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
	if _, err := bump(&Config{Packages: map[string]string{"a/a": "1"}}, "http://127.0.0.1:0"); err == nil {
		t.Error("unreachable packagist: no error")
	}
}
