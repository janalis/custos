// Package deprecatedfeature implements the native DeprecatedFeature inspection.
// It reports code using what PHP deprecates at the target version.
package deprecatedfeature

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
)

type rule struct{}

// New returns the DeprecatedFeature inspection.
func New() analysis.Rule { return rule{} }

func (rule) ID() string { return "DeprecatedFeature" }
func (rule) Semantic()  {}

// PartialCalls: a partial application refers to a deprecated function too.
func (rule) PartialCalls() {}

func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{
		syntax.KFuncCall, syntax.KMethodCall, syntax.KStaticCall, syntax.KNew, syntax.KConstFetch, syntax.KClassConstFetch,
		syntax.KReturn, syntax.KYield, syntax.KYieldFrom, syntax.KMethod,
		syntax.KClassConst, syntax.KClassLike, syntax.KFunction, syntax.KConstStmt, syntax.KUse,
	}
}
