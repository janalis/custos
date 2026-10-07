package conformance

import (
	"encoding/json"
	"strconv"
	"strings"

	"custos/internal/analysis"
	"custos/internal/fix"
	"custos/internal/index"
	"custos/internal/meta"
	"custos/internal/phpver"
	"custos/internal/rules"
	"custos/internal/stubs"
	"custos/internal/syntax"
)

// eaDefaultPHP is the language level EA tests run at when they set none
// (PhpStorm's test default). Fixture evidence pins it to 5.6: ini options
// removed in 7.0 are reported as deprecated, argument unpacking (5.6) is
// suggested, and visibility on class constants (7.1) is not.
const eaDefaultPHP = phpver.PHP56

func init() { engine = &adapter{registered: rules.All()} }

type adapter struct {
	registered []analysis.Rule
}

func (a *adapter) Supports(ids []string) bool {
	if len(ids) == 0 {
		return false
	}
	for _, id := range ids {
		found := false
		for _, r := range a.registered {
			if r.ID() == id {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// target resolves the rule-gating version: explicit case level, else the EA
// default for EA cases or the custos default for own fixtures.
func target(req Request) phpver.Version {
	if req.PHP != "" {
		if v, err := phpver.Parse(req.PHP); err == nil {
			return v
		}
	}
	if strings.HasPrefix(req.Path, "testData/") {
		return eaDefaultPHP
	}
	return phpver.Default
}

func (a *adapter) engineFor(req Request) (*analysis.Engine, error) {
	cfg := analysis.Config{PHP: target(req), Only: req.Rules, Rules: map[string]analysis.RuleConfig{}}
	if req.ComparisonStyle == "yoda" {
		cfg.ComparisonStyle = analysis.StyleYoda
	}
	for _, id := range req.Rules {
		rc := analysis.RuleConfig{Options: map[string]any{}}
		m, _ := meta.Lookup(id)
		for key, raw := range req.Options {
			ruleID, opt, ok := strings.Cut(key, ".")
			if !ok || ruleID != id {
				continue
			}
			rc.Options[opt] = convertOption(m, opt, raw)
		}
		var calls []string
		for _, c := range req.Calls {
			if strings.HasPrefix(c, id+".") {
				calls = append(calls, strings.TrimPrefix(c, id+"."))
			}
		}
		if len(calls) > 0 {
			rc.Options["@calls"] = calls
		}
		cfg.Rules[id] = rc
	}
	e, err := analysis.NewEngine(a.registered, cfg)
	if err != nil {
		return nil, err
	}
	if len(req.Companions) > 0 {
		ix := index.New(stubs.Index())
		for path, src := range req.Companions {
			ix.Add(index.Extract(syntax.Parse(path, src, parseOpts())))
		}
		e.SetIndex(ix)
	}
	return e, nil
}

// convertOption turns a raw Java literal into the option's Go type.
func convertOption(m *meta.Rule, name, raw string) any {
	raw = strings.TrimSpace(raw)
	typ := ""
	if m != nil {
		for _, o := range m.Options {
			if o.Name == name {
				typ = o.Type
			}
		}
	}
	switch typ {
	case "bool":
		return raw == "true"
	case "int":
		n, _ := strconv.Atoi(raw)
		return n
	case "enum":
		if i := strings.LastIndexByte(raw, '.'); i >= 0 {
			return raw[i+1:]
		}
	case "string":
		return strings.Trim(raw, `"`)
	case "list":
		var list []any
		if json.Unmarshal([]byte(raw), &list) == nil {
			return list
		}
	}
	return raw
}

func parseOpts() syntax.Options {
	// PhpStorm parses every construct at any language level, recognising `<?`.
	return syntax.Options{Version: phpver.Max, Permissive: true, ShortOpenTag: true}
}

func (a *adapter) Analyze(req Request) ([]Finding, error) {
	e, err := a.engineFor(req)
	if err != nil {
		return nil, err
	}
	f := syntax.Parse(req.Path, req.Source, parseOpts())
	var out []Finding
	for _, fd := range e.Analyze(f) {
		out = append(out, Finding{Rule: fd.Rule, Severity: fd.Severity, Message: fd.Message, Start: int(fd.Span.Start), End: int(fd.Span.End)})
	}
	return out, nil
}

func (a *adapter) Fix(req Request) ([]byte, error) {
	e, err := a.engineFor(req)
	if err != nil {
		return nil, err
	}
	res := fix.FixSource(e, req.Path, req.Source, fix.Options{Parse: parseOpts(), SinglePass: true})
	return res.Source, nil
}
