// Package opensslencryptionstatususedasciphertext implements OpenSslEncryptionStatusUsedAsCiphertext.
package opensslencryptionstatususedasciphertext

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use the encryption output as ciphertext."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "OpenSslEncryptionStatusUsedAsCiphertext" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall, syntax.KEcho} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckOpenSslEncryptionStatusUsedAsCiphertext(ctx, n, message)
}
