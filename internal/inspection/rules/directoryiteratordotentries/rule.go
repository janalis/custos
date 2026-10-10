// Package directoryiteratordotentries implements the native DirectoryIteratorDotEntries inspection.
package directoryiteratordotentries

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Skip dot entries before modifying directory entries."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DirectoryIteratorDotEntries" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckDirectoryIteratorDotEntries(ctx, n, message)
}
