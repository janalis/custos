// Package aeadauthenticationtagdiscarded implements the native AeadAuthenticationTagDiscarded inspection.
package aeadauthenticationtagdiscarded

import (
	"strings"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
)

const message = "Capture the authenticated encryption tag."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "AeadAuthenticationTagDiscarded" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP.Below(phpversion.PHP71) {
		return
	}
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "openssl_encrypt") {
		return
	}
	cipher, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 1, "cipher_algo"))
	if !known {
		return
	}
	switch strings.ToLower(cipher) {
	case "aes-128-gcm", "aes-192-gcm", "aes-256-gcm", "aes-128-ccm", "aes-192-ccm", "aes-256-ccm":
	default:
		return
	}
	if semanticquery.CallArgument(call.Args, 5, "tag") == nil {
		ctx.ReportNode(call, message)
	}
}
