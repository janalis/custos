package flow

import (
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
)

// sinkSummaryName accepts method bodies only when inheritance cannot replace
// them. Absence of a known override is insufficient in an open project.
func (e *Env) sinkSummaryName(at syntax.Node) string {
	if _, function := at.(*syntax.FuncCall); function {
		return e.summaryName(at)
	}
	receiver := ""
	switch n := at.(type) {
	case *syntax.MethodCall:
		classes := e.types.TypeOf(n.Var).Classes()
		if len(classes) != 1 {
			return ""
		}
		receiver = classes[0]
	case *syntax.StaticCall:
		class, ok := n.Class.(*syntax.Name)
		if !ok {
			return ""
		}
		receiver = e.named(class)
	default:
		return ""
	}
	m := e.method(at)
	if m != nil {
		e.contract(sinkClassKey(m.Class))
	}
	// Protected accessibility also depends on the lexical class's ancestry.
	within := e.types.ClassFQN(syntax.EnclosingClass(at))
	for steps := 0; within != "" && steps < MaxDepth*4; steps++ {
		e.contract(sinkClassKey(within))
		decls := e.types.Index.ClassDecls(within, e.types.PHP)
		if len(decls) != 1 {
			return ""
		}
		within = decls[0].Parent
	}
	exact := m != nil && (m.Final || m.Visibility == index.Private)
	for steps := 0; steps < MaxDepth*4; steps++ {
		meta, _ := e.contract(sinkClassKey(receiver))
		decls := e.types.Index.ClassDecls(receiver, e.types.PHP)
		if len(decls) != 1 || decls[0].Kind == syntax.KindTrait {
			return ""
		}
		c := decls[0]
		if m == nil {
			if c.Parent == "" {
				return ""
			}
			receiver = c.Parent
			continue
		}
		if steps == 0 {
			exact = exact || c.Final
		}
		if strings.EqualFold(strings.TrimPrefix(c.FQN, `\`), strings.TrimPrefix(m.Class, `\`)) {
			if _, ok := at.(*syntax.StaticCall); ok && !m.Static {
				return ""
			}
			if m.Magic || !e.visible(at, m) || !exact || meta.dispatch == nil || meta.dispatch.Methods[strings.ToLower(m.Name)].Count != 1 {
				return ""
			}
			name := strings.ToLower(m.Class + "::" + m.Name)
			if summary, ok := e.contract(name); ok && summary.Fixed {
				return name
			}
			return ""
		}
		if c.Parent == "" {
			return ""
		}
		receiver = c.Parent
	}
	return ""
}

// Reserved keys cannot collide with PHP callable identifiers. Class metadata
// participates in the existing snapshot dependency, overlay and equality paths.
func sinkClassKey(name string) string {
	return "\x00sink-class:" + names.Fold(strings.TrimPrefix(name, `\`))
}

type sinkClass struct {
	Final     bool
	Kind      syntax.ClassKind
	Modifiers string
	Extends   []string
	Traits    []string
	Methods   map[string]sinkMethod
}
type sinkMethod struct {
	Count      int
	Modifiers  string
	Parameters []index.Param
}

func describeSinkClass(n *syntax.ClassLike, types *infer.Env) *sinkClass {
	c := &sinkClass{Final: n.Modifiers.Has(syntax.TFinal), Kind: n.ClassKind, Methods: map[string]sinkMethod{}}
	for _, m := range n.Modifiers {
		c.Modifiers += m.Kind.String() + " "
	}
	for _, p := range n.Extends {
		c.Extends = append(c.Extends, types.Names.Class(p.Value, p.Span().Start))
	}
	for _, member := range n.Members {
		switch m := member.(type) {
		case *syntax.Method:
			if m.Name != nil {
				key := strings.ToLower(m.Name.Value)
				v := c.Methods[key]
				v.Count++
				if declared := types.Index.FindMethod(types.ClassFQN(n), m.Name.Value, types.PHP); declared != nil {
					v.Parameters = append([]index.Param(nil), declared.Params...)
				}
				for _, mod := range m.Modifiers {
					v.Modifiers += mod.Kind.String() + " "
				}
				c.Methods[key] = v
			}
		case *syntax.TraitUse:
			for _, name := range m.Traits {
				c.Traits = append(c.Traits, types.Names.Class(name.Value, name.Span().Start))
			}
		}
	}
	return c
}
