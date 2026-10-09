package phpdoc

import "strings"

// EffectiveParams selects parameter types, preferring phpstan-param over
// psalm-param over param independently for each named parameter. Empty types
// are ignored. Within one dialect the last declaration wins, as with Params.
// Results follow the first occurrence of each selected parameter in the doc.
func (d *Doc) EffectiveParams() []Param {
	return d.effectiveTargets("param", true)
}

// EffectiveVars selects variable types, preferring phpstan-var over psalm-var
// over var independently for each variable (including unnamed annotations).
// Empty types are ignored; the first declaration within a dialect wins.
// Results follow the first occurrence of each selected variable in the doc.
func (d *Doc) EffectiveVars() []Param {
	return d.effectiveTargets("var", false)
}

func annotationRank(name, base string) int {
	switch name {
	case "phpstan-" + base:
		return 3
	case "psalm-" + base:
		return 2
	case base:
		return 1
	}
	return 0
}

func typedTarget(text string, parameter bool) Param {
	typ, rest := SplitType(text)
	if strings.HasPrefix(typ, "$") {
		name := VarName(typ)
		if parameter {
			return Param{Type: trailingType(rest), Name: name}
		}
		typ, _ = SplitType(rest)
		return Param{Type: typ, Name: name}
	}
	return Param{Type: typ, Name: VarName(rest)}
}

func (d *Doc) effectiveTargets(base string, parameter bool) []Param {
	var out []Param
	positions := make(map[string]int)
	var ranks []int
	for _, tag := range d.Tags {
		rank := annotationRank(tag.Name, base)
		if rank == 0 {
			continue
		}
		p := typedTarget(tag.Text, parameter)
		if p.Type == "" || parameter && p.Name == "" {
			continue
		}
		if i, exists := positions[p.Name]; exists {
			if rank > ranks[i] || parameter && rank == ranks[i] {
				out[i], ranks[i] = p, rank
			}
			continue
		}
		positions[p.Name] = len(out)
		out = append(out, p)
		ranks = append(ranks, rank)
	}
	return out
}

// EffectiveReturnType selects the first nonempty return type in the highest
// available dialect: phpstan-return, psalm-return, then return. Description-only
// annotations are ignored using the same heuristic as ReturnType.
func (d *Doc) EffectiveReturnType() string {
	var out string
	selected := 0
	for _, tag := range d.Tags {
		rank := annotationRank(tag.Name, "return")
		if rank <= selected {
			continue
		}
		typ, desc := SplitType(tag.Text)
		if typ == "" || desc != "" && proseWord[strings.ToLower(typ)] {
			continue
		}
		out, selected = typ, rank
	}
	return out
}

// EffectiveVarType selects the first applicable nonempty variable annotation
// in the highest available dialect. An unnamed annotation applies to any name;
// an empty name selects the first variable annotation in that dialect.
func (d *Doc) EffectiveVarType(name string) string {
	var out string
	selected := 0
	for _, tag := range d.Tags {
		rank := annotationRank(tag.Name, "var")
		if rank <= selected {
			continue
		}
		p := typedTarget(tag.Text, false)
		if p.Type == "" || name != "" && p.Name != "" && p.Name != name {
			continue
		}
		out, selected = p.Type, rank
	}
	return out
}
