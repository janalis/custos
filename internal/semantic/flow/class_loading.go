package flow

import (
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// className reports `X::class`: the name of a class, as text.
func className(n *syntax.ClassConstFetch) bool {
	name, ok := n.Name.(*syntax.Identifier)
	return ok && strings.EqualFold(name.Value, "class")
}

// firstArg returns the first argument of a call written in order.
func firstArg(list *syntax.ArgList) (*syntax.Arg, bool) {
	if list == nil || len(list.Args) == 0 {
		return nil, false
	}
	arg, plain := list.Args[0].(*syntax.Arg)
	return arg, plain && !arg.Unpack && arg.Name == nil
}

// computedDefinesKey caches syntax only; no semantic option or index state.
type computedDefinesKey struct{}

// computedDefines reports a call of define in the file whose name is not
// literal text.
func (e *Env) computedDefines() bool {
	return e.file.Memo(computedDefinesKey{}, func() any {
		found := false
		syntax.InspectFile(e.file, func(n syntax.Node) bool {
			if found {
				return false
			}
			call, ok := n.(*syntax.FuncCall)
			if name, builtin := e.callName(n); ok && builtin && name == "define" {
				found = true
				if arg, plain := firstArg(call.Args); plain {
					_, literal := literalText(arg.Value)
					found = !literal
				}
			}
			return !found
		})
		return found
	}).(bool)
}

// selfContained reports a class constant whose value names no other class:
// it is written without `::`, or only with `self::` before constants of the
// declaring class that are so themselves, a bounded number of steps deep.
func (e *Env) selfContained(constant *index.ClassConst, depth int) bool {
	if constant.Case || constant.Value == "" || depth == 0 {
		return false
	}
	named := func(c byte) bool {
		return c == '_' || c == '\\' || c >= '0' && c <= '9' || c|0x20 >= 'a' && c|0x20 <= 'z' || c >= 0x80
	}
	for rest := constant.Value; ; {
		at := strings.Index(rest, "::")
		if at < 0 {
			return true
		}
		from, to := at, at+2
		for from > 0 && named(rest[from-1]) {
			from--
		}
		for to < len(rest) && named(rest[to]) {
			to++
		}
		other := e.types.Index.FindConst(constant.Class, rest[at+2:to], e.types.PHP)
		if !strings.EqualFold(rest[from:at], "self") || other == nil || !strings.EqualFold(other.Class, constant.Class) || !e.selfContained(other, depth-1) {
			return false
		}
		rest = rest[to:]
	}
}

// rooted reports a class of the project whose ancestors are all known: each
// class up to the top is declared once, and so is each trait they use.
func (e *Env) rooted(fqn string) bool {
	ix := e.types.Index
	for steps := 0; steps < MaxDepth*4; steps++ {
		class := ix.Class(fqn, e.types.PHP)
		if class == nil || ix.ClassCount(fqn) != 1 {
			return false
		}
		for _, trait := range class.Traits {
			if ix.ClassCount(trait) != 1 {
				return false
			}
		}
		if fqn = class.Parent; fqn == "" {
			return true
		}
	}
	return false
}

// backed reports a property read that runs no code: `$this->name` in a
// method of a class that declares the property itself, without a hook, where
// no `__get` method can stand in for it. A class extending this one can only
// make the read run code by declaring `__get` (for a property removed with
// `unset`) or, unless the property is private, by declaring it again with a
// hook: the class is final, or no class of the project extending it has such
// a method and the property is private. An object PHP 8.4 made lazy runs its
// initializer on the first read: that is not followed.
func (e *Env) backed(n *syntax.PropertyFetch) bool {
	this, direct := n.Var.(*syntax.Variable)
	name, written := n.Name.(*syntax.Identifier)
	_, method := syntax.EnclosingVariableScope(n).(*syntax.Method)
	within := syntax.EnclosingClass(n)
	if !direct || this.Name != "this" || this.NameExpr != nil || !written || !method || within == nil || within.ClassKind != syntax.KindClass {
		return false
	}
	at := [2]string{e.types.ClassFQN(within), name.Value}
	if known, read := e.properties[at]; read {
		// The answer may have read whether a subclass declares __get.
		e.depend(known.reads)
		return known.holds
	}
	if e.properties == nil {
		e.properties = map[[2]string]cachedAnswer{}
	}
	reads := e.begin()
	holds := e.declares(at[0], at[1])
	e.end(reads)
	e.properties[at] = cachedAnswer{holds: holds, reads: reads}
	return holds
}

// declares is backed for a class and the name of a property.
func (e *Env) declares(fqn, name string) bool {
	ix := e.types.Index
	property := ix.FindProperty(fqn, name, e.types.PHP)
	if property == nil || property.Magic || property.Static || property.Hooked || property.ReadsRunCode || !strings.EqualFold(strings.TrimPrefix(property.Class, `\`), strings.TrimPrefix(fqn, `\`)) {
		return false
	}
	if !e.rooted(fqn) || ix.FindMethod(fqn, "__get", e.types.PHP) != nil {
		return false
	}
	return ix.Class(fqn, e.types.PHP).Final || property.Visibility == index.Private && !e.overrides(fqn, "__get", "")
}

// loaded reports a class constant read that loads no class: a constant of
// the enclosing class, named `self`, `parent` or by its own name, or
// `static` when no class can declare it again, whose value is written
// without a constant of another class (see selfContained). An enum case may
// name one.
func (e *Env) loaded(n *syntax.ClassConstFetch) bool {
	class, named := n.Class.(*syntax.Name)
	name, written := n.Name.(*syntax.Identifier)
	within := syntax.EnclosingClass(n)
	if !named || !written || within == nil || within.ClassKind == syntax.KindEnum {
		return false
	}
	own := e.types.ClassFQN(within)
	fqn := e.named(class)
	switch strings.ToLower(class.Value) {
	case "self", "parent", "static":
	default:
		if !strings.EqualFold(strings.TrimPrefix(fqn, `\`), strings.TrimPrefix(own, `\`)) {
			return false
		}
	}
	constant := e.types.Index.FindConst(fqn, name.Value, e.types.PHP)
	if constant == nil || !e.selfContained(constant, MaxDepth) || !e.rooted(own) {
		return false
	}
	if strings.EqualFold(class.Value, "static") {
		declared := e.types.Index.Class(own, e.types.PHP)
		return constant.Final || declared.Final
	}
	return true
}
