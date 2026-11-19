// Package ast exposes the PHP syntax tree that custom rules written in Go
// inspect: node types and kinds, token kinds and navigation helpers. Nodes
// carry byte offsets into the file's source.
//
// The API is unstable until custos 1.0: node shapes follow the parser.
package ast

import "github.com/janalis/custos/internal/php/syntax"

// Navigation helpers.
var (
	// Children calls fn for each direct child of n, in source order.
	Children = syntax.Children
	// Inspect walks the tree below n depth first; fn returning false skips
	// a node's children.
	Inspect = syntax.Inspect
	// UnwrapParens returns e without its enclosing parentheses.
	UnwrapParens = syntax.UnwrapParens
	// EnclosingClass returns the class-like declaration around n, or nil.
	EnclosingClass = syntax.EnclosingClass
	// EnclosingFuncLike returns the function, method, closure or arrow
	// function around n, or nil.
	EnclosingFuncLike = syntax.EnclosingFuncLike
	// IsNullConst reports whether e is the constant null.
	IsNullConst = syntax.IsNullConst
	// LiteralStringValue returns the value of a string literal.
	LiteralStringValue = syntax.LiteralStringValue
)
