// Package rules is the registry of all implemented rules.
package rules

import (
	"custos/internal/analysis"
	"custos/internal/rules/architecture"
	"custos/internal/rules/codestyle"
	"custos/internal/rules/compatibility"
	"custos/internal/rules/confusing"
	"custos/internal/rules/controlflow"
	"custos/internal/rules/langmigration"
	"custos/internal/rules/performance"
	"custos/internal/rules/phpunit"
	"custos/internal/rules/probablebugs"
	"custos/internal/rules/security"
	"custos/internal/rules/typecompat"
	"custos/internal/rules/unused"
)

// All returns every implemented rule.
func All() []analysis.Rule {
	var out []analysis.Rule
	for _, group := range [][]analysis.Rule{
		architecture.Rules(), codestyle.Rules(), compatibility.Rules(), confusing.Rules(),
		controlflow.Rules(), langmigration.Rules(), performance.Rules(), phpunit.Rules(),
		probablebugs.Rules(), security.Rules(), typecompat.Rules(), unused.Rules(),
	} {
		out = append(out, group...)
	}
	return out
}
