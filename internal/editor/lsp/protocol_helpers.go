package lsp

import (
	"encoding/json"
	"net/url"
	"path/filepath"
)

func mustJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func uriToPath(uri string) string {
	u, err := url.Parse(uri)
	if err != nil || u.Scheme != "file" {
		return uri
	}
	p := u.Path
	// file:///C:/dir → C:/dir (Windows drive letters).
	if len(p) >= 3 && p[0] == '/' && p[2] == ':' && ((p[1] >= 'a' && p[1] <= 'z') || (p[1] >= 'A' && p[1] <= 'Z')) {
		p = p[1:]
	}
	return filepath.FromSlash(p)
}
