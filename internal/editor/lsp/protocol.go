package lsp

import "encoding/json"

// Minimal LSP protocol types used by the server.

type position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type lspRange struct {
	Start position `json:"start"`
	End   position `json:"end"`
}

type textDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

type versionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version *int   `json:"version"`
}

type initializeParams struct {
	RootURI               string          `json:"rootUri"`
	RootPath              string          `json:"rootPath"`
	InitializationOptions json.RawMessage `json:"initializationOptions"`
	WorkspaceFolders      []struct {
		URI string `json:"uri"`
	} `json:"workspaceFolders"`
	Capabilities struct {
		Workspace struct {
			DidChangeWatchedFiles struct {
				DynamicRegistration bool `json:"dynamicRegistration"`
			} `json:"didChangeWatchedFiles"`
		} `json:"workspace"`
		TextDocument struct {
			CodeAction struct {
				ResolveSupport *struct {
					Properties []string `json:"properties"`
				} `json:"resolveSupport"`
			} `json:"codeAction"`
		} `json:"textDocument"`
	} `json:"capabilities"`
}

type didOpenParams struct {
	TextDocument textDocumentItem `json:"textDocument"`
}

type didChangeParams struct {
	TextDocument   versionedTextDocumentIdentifierInt `json:"textDocument"`
	ContentChanges []struct {
		Range *lspRange `json:"range"`
		Text  string    `json:"text"`
	} `json:"contentChanges"`
}

type versionedTextDocumentIdentifierInt struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

type protocolDiagnostic struct {
	Range    lspRange `json:"range"`
	Severity int      `json:"severity"`
	Code     string   `json:"code,omitempty"`
	Source   string   `json:"source,omitempty"`
	Message  string   `json:"message"`
	Tags     []int    `json:"tags,omitempty"`
	// Data identifies the finding (rule + byte span); clients send it back
	// in the code-action context so fixes can be matched exactly.
	Data *diagData `json:"data,omitempty"`
}

type diagData struct {
	Rule  string `json:"rule"`
	Start uint32 `json:"start"`
	End   uint32 `json:"end"`
}

type publishDiagnosticsParams struct {
	URI         string               `json:"uri"`
	Version     *int                 `json:"version,omitempty"`
	Diagnostics []protocolDiagnostic `json:"diagnostics"`
}

type codeActionParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
	Range   lspRange `json:"range"`
	Context struct {
		Diagnostics []protocolDiagnostic `json:"diagnostics"`
		Only        []string             `json:"only"`
	} `json:"context"`
}

type textEdit struct {
	Range   lspRange `json:"range"`
	NewText string   `json:"newText"`
}

type textDocumentEdit struct {
	TextDocument versionedTextDocumentIdentifier `json:"textDocument"`
	Edits        []textEdit                      `json:"edits"`
}

type workspaceEdit struct {
	DocumentChanges []textDocumentEdit `json:"documentChanges"`
}

type codeAction struct {
	Title       string               `json:"title"`
	Kind        string               `json:"kind,omitempty"`
	Diagnostics []protocolDiagnostic `json:"diagnostics,omitempty"`
	IsPreferred bool                 `json:"isPreferred,omitempty"`
	Edit        *workspaceEdit       `json:"edit,omitempty"`
	Data        json.RawMessage      `json:"data,omitempty"`
}

type executeCommandParams struct {
	Command   string            `json:"command"`
	Arguments []json.RawMessage `json:"arguments"`
}
