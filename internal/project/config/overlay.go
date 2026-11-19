package config

import (
	"maps"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/infer"
)

// Overlay applies settings a client supplies (an editor's) on top of the
// project configuration base. What to analyze (paths, exclusions, the
// baseline) stays the project's; a rule entry overrides the project's
// field by field.
func Overlay(base *Config, over File) (*Config, error) {
	f := File{Flow: &FlowSettings{}, PHP: over.PHP, ComparisonStyle: over.ComparisonStyle, TypeCheck: over.TypeCheck, ShortOpenTag: over.ShortOpenTag, Rules: over.Rules}
	if f.PHP == "" {
		f.PHP = base.PHP.String()
	}
	if f.ComparisonStyle == "" && base.ComparisonStyle == analysis.StyleYoda {
		f.ComparisonStyle = "yoda"
	}
	if over.Flow != nil {
		*f.Flow = *over.Flow
	}
	if f.Flow.Assertions == nil {
		f.Flow.Assertions = &base.Flow.Assertions
	}
	if f.Flow.Order == nil {
		f.Flow.Order = &base.FlowOrder
	}
	if f.ShortOpenTag == nil {
		f.ShortOpenTag = &base.ShortOpenTag
	}
	merged, err := Resolve(base.Root, f)
	if err != nil {
		return nil, err
	}
	if over.TypeCheck == "" {
		merged.TypeCheck = base.TypeCheck
	}
	// What to analyze stays the project's choice.
	merged.Paths, merged.Exclude, merged.Baseline = base.Paths, base.Exclude, base.Baseline
	// A client rule entry overrides the project's field by field.
	for id, rc := range base.Rules {
		over, ok := merged.Rules[id]
		if !ok {
			merged.Rules[id] = rc
			continue
		}
		if over.Enabled == nil {
			over.Enabled = rc.Enabled
		}
		if over.Severity == "" {
			over.Severity = rc.Severity
		}
		opts := maps.Clone(rc.Options)
		if opts == nil {
			opts = map[string]any{}
		}
		maps.Copy(opts, over.Options)
		over.Options = opts
		merged.Rules[id] = over
	}
	return merged, nil
}

// Parse returns the options files are parsed with: the target PHP version
// and the short open tag setting.
func (c *Config) Parse() syntax.Options {
	return syntax.Options{Version: c.PHP, ShortOpenTag: c.ShortOpenTag}
}

// Infer returns the options type inference runs with.
func (c *Config) Infer() infer.Options {
	return infer.Options{Assertions: c.Flow.Assertions}
}
