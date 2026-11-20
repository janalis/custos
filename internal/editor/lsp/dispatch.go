package lsp

import (
	"encoding/json"
	"path/filepath"
	"strings"
)

func (s *Server) handle(m *message) (result any, rerr *rpcError) {
	// A crashing rule must never take the server down; a request that
	// crashed gets an error, not a null success.
	defer s.guard("handling "+m.Method, func() {
		result, rerr = nil, &rpcError{Code: codeInternalError, Message: "custos crashed handling " + m.Method}
	})
	switch m.Method {
	case "initialize":
		return s.initialize(m.Params)
	case "initialized":
		if s.watchDynamic {
			// Ask the client to report PHP file changes so the project index
			// follows edits made outside open buffers (git checkout, codegen…),
			// baseline changes so regenerating it takes effect, and Composer
			// metadata changes so the index world follows installs.
			// JSON covers Composer's metadata, custos.json and the custom rule
			// files, including those the configuration lists later.
			// .editorconfig files and the other formatters' configuration
			// set the format settings.
			watchers := []any{map[string]any{"globPattern": "**/*.php"}, map[string]any{"globPattern": "**/*.json"}, map[string]any{"globPattern": "**/.editorconfig"}, map[string]any{"globPattern": "**/.php-cs-fixer{,.dist}.php"}, map[string]any{"globPattern": "**/mago{,.dist}.toml"}}
			s.mu.Lock()
			if p := baselinePath(s.cfg); p != "" {
				watchers = append(watchers, map[string]any{"globPattern": "**/" + filepath.Base(p)})
			}
			for _, p := range configInputs(s.cfg, s.root) {
				if !strings.HasSuffix(p, ".json") {
					watchers = append(watchers, map[string]any{"globPattern": "**/" + filepath.Base(p)})
				}
			}
			s.mu.Unlock()
			_ = s.c.request("client/registerCapability", map[string]any{"registrations": []any{map[string]any{
				"id": "custos-php-watch", "method": "workspace/didChangeWatchedFiles",
				"registerOptions": map[string]any{"watchers": watchers},
			}}})
		}
		if s.beginIndexing() {
			go s.buildIndex()
		}
		return nil, nil
	case "workspace/didChangeWatchedFiles":
		var p struct {
			Changes []fileChange `json:"changes"`
		}
		_ = json.Unmarshal(m.Params, &p)
		go func() {
			defer s.guard("file changes", nil)
			s.watchedFilesChanged(p.Changes)
		}()
		return nil, nil
	case "textDocument/didSave":
		var p struct {
			TextDocument struct {
				URI string `json:"uri"`
			} `json:"textDocument"`
		}
		_ = json.Unmarshal(m.Params, &p)
		go func() {
			defer s.guard("save", nil)
			// The saved file's symbols may change what other open documents
			// report (e.g. a method added or removed), so they are re-analyzed.
			s.reflow(p.TextDocument.URI)
		}()
		return nil, nil
	case "$/cancelRequest", "$/setTrace":
		return nil, nil
	case "shutdown":
		s.shutdown = true
		return nil, nil
	case "textDocument/didOpen":
		var p didOpenParams
		if err := json.Unmarshal(m.Params, &p); err != nil {
			return nil, &rpcError{Code: codeInvalidParams, Message: err.Error()}
		}
		s.open(p)
		return nil, nil
	case "textDocument/didChange":
		var p didChangeParams
		if err := json.Unmarshal(m.Params, &p); err != nil {
			return nil, &rpcError{Code: codeInvalidParams, Message: err.Error()}
		}
		s.change(p)
		return nil, nil
	case "textDocument/didClose":
		var p struct {
			TextDocument struct {
				URI string `json:"uri"`
			} `json:"textDocument"`
		}
		_ = json.Unmarshal(m.Params, &p)
		s.close(p.TextDocument.URI)
		return nil, nil
	case "workspace/didChangeConfiguration":
		var p struct {
			Settings json.RawMessage `json:"settings"`
		}
		_ = json.Unmarshal(m.Params, &p)
		if len(p.Settings) > 0 && string(p.Settings) != "null" {
			s.mu.Lock()
			s.initOpts = p.Settings
			s.mu.Unlock()
		}
		if err := s.configure(); err != nil {
			s.logf("configuration: %v", err)
		}
		s.reanalyzeAll()
		return nil, nil
	case "textDocument/codeAction":
		var p codeActionParams
		if err := json.Unmarshal(m.Params, &p); err != nil {
			return nil, &rpcError{Code: codeInvalidParams, Message: err.Error()}
		}
		return s.codeActions(p), nil
	case "codeAction/resolve":
		var a codeAction
		if err := json.Unmarshal(m.Params, &a); err != nil {
			return nil, &rpcError{Code: codeInvalidParams, Message: err.Error()}
		}
		return s.resolveAction(a)
	case "textDocument/formatting", "textDocument/rangeFormatting":
		var p formattingParams
		if err := json.Unmarshal(m.Params, &p); err != nil {
			return nil, &rpcError{Code: codeInvalidParams, Message: err.Error()}
		}
		if m.Method == "textDocument/rangeFormatting" && p.Range == nil {
			return nil, &rpcError{Code: codeInvalidParams, Message: "rangeFormatting needs a range"}
		}
		return s.formatting(p), nil
	case "workspace/executeCommand":
		var p executeCommandParams
		if err := json.Unmarshal(m.Params, &p); err != nil {
			return nil, &rpcError{Code: codeInvalidParams, Message: err.Error()}
		}
		return s.executeCommand(p)
	}
	if m.ID != nil {
		return nil, &rpcError{Code: codeMethodNotFound, Message: "method not found: " + m.Method}
	}
	return nil, nil
}
