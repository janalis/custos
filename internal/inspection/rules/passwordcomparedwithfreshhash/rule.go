// Package passwordcomparedwithfreshhash implements the native PasswordComparedWithFreshHash inspection.
package passwordcomparedwithfreshhash

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Verify the password against the stored hash."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PasswordComparedWithFreshHash" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	// A comparison with false, null, a number, an array or the empty string
	// tests whether hashing succeeded; no stored hash takes part in it.
	if binary := n.(*syntax.Binary); notStoredHash(ctx, binary.Left) || notStoredHash(ctx, binary.Right) {
		return
	}
	semanticquery.CheckPasswordComparedWithFreshHash(ctx, n, message)
}

// notStoredHash reports whether e is known to hold a value that cannot be a
// stored password hash: a boolean, null, a number, an array or "".
func notStoredHash(ctx *analysis.Context, e syntax.Expr) bool {
	switch v := semanticquery.NativeValue(ctx, e).(type) {
	case *syntax.ConstFetch:
		switch strings.ToLower(strings.TrimPrefix(v.Name.Value, `\`)) {
		case "false", "null", "true":
			return true
		}
	case *syntax.Literal:
		if v.LitKind != syntax.LitString {
			return true
		}
		text, known := semanticquery.NativeString(ctx, v)
		return known && text == ""
	case *syntax.Array:
		return true
	}
	return false
}
