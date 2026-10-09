package infer

import (
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

// literalMethodCallback resolves callback syntax without inventing the arguments
// supplied later by a callback consumer. Runtime strings name absolute classes.
func (e *Env) literalMethodCallback(x syntax.Expr) (types.Type, bool) {
	x = syntax.UnwrapParens(x)
	if lit, ok := x.(*syntax.Literal); ok && lit.LitKind == syntax.LitString {
		value, valid := plainString(lit.Raw)
		if !valid || !strings.Contains(value, ":") {
			return types.Unknown, false
		}
		cls, method, found := strings.Cut(value, "::")
		if !found || cls == "" || method == "" || strings.Contains(method, ":") {
			return types.Unknown, true
		}
		return e.callbackMethodReturn(strings.TrimPrefix(cls, `\`), method, true, nil), true
	}
	arr, ok := x.(*syntax.Array)
	if !ok {
		return types.Unknown, false
	}
	if len(arr.Items) != 2 {
		return types.Unknown, true
	}
	for _, item := range arr.Items {
		if item == nil || item.Key != nil || item.ByRef || item.Unpack || item.Value == nil {
			return types.Unknown, true
		}
	}
	methodLit, ok := syntax.UnwrapParens(arr.Items[1].Value).(*syntax.Literal)
	if !ok || methodLit.LitKind != syntax.LitString {
		return types.Unknown, true
	}
	method, valid := plainString(methodLit.Raw)
	if !valid || method == "" {
		return types.Unknown, true
	}
	receiver := syntax.UnwrapParens(arr.Items[0].Value)
	if lit, ok := receiver.(*syntax.Literal); ok && lit.LitKind == syntax.LitString {
		cls, valid := plainString(lit.Raw)
		if !valid || cls == "" {
			return types.Unknown, true
		}
		return e.callbackMethodReturn(strings.TrimPrefix(cls, `\`), method, true, nil), true
	}
	if fetch, ok := receiver.(*syntax.ClassConstFetch); ok {
		id, named := fetch.Name.(*syntax.Identifier)
		_, literalClass := fetch.Class.(*syntax.Name)
		if !named || !strings.EqualFold(id.Value, "class") || !literalClass {
			return types.Unknown, true
		}
		cls := e.classRef(fetch.Class)
		if cls == "" {
			return types.Unknown, true
		}
		return e.callbackMethodReturn(cls, method, true, nil), true
	}
	recv := e.TypeOf(receiver)
	if recv.IsUnknown() {
		return types.Unknown, true
	}
	var returns []types.Type
	for _, atom := range recv.Atoms() {
		if !strings.HasPrefix(atom, `\`) || strings.HasSuffix(atom, "[]") {
			return types.Unknown, true
		}
		ret := e.callbackMethodReturn(strings.TrimPrefix(atom, `\`), method, false, recv.TypeArgs(atom))
		if ret.IsUnknown() {
			return types.Unknown, true
		}
		returns = append(returns, ret)
	}
	return types.Union(returns...), true
}

func (e *Env) callbackMethodReturn(cls, method string, static bool, args []types.Type) types.Type {
	m := e.Index.FindMethod(cls, method, e.PHP)
	if m == nil || m.Visibility != index.Public || (static && !m.Static) {
		return types.Unknown
	}
	// Method templates and conditional returns require the eventual arguments.
	if m.Tpl != nil || m.CondReturn != "" {
		return types.Unknown
	}
	return voidAsNull(e.methodReturn(cls, method, !static, cls, args, nil))
}
