package flow

import (
	"strings"

	"custos/internal/php/syntax"
)

// evalArrayDimFetch reads an element of an array, or an offset of an object.
func (e *Env) evalArrayDimFetch(n *syntax.ArrayDimFetch, f *frame, r *scopeResult, depth int) Value {
	if name := assigned(n); external(name) {
		e.eval(n.Var, f, r, depth+1)
		e.eval(n.Dim, f, r, depth+1)
		return e.request(name, n, f)
	}
	v := e.eval(n.Var, f, r, depth+1)
	if conditional(n.Var) {
		indexed := clone(f)
		e.eval(n.Dim, &indexed, r, depth+1)
		*f = rejoin(*f, indexed, r)
	} else {
		e.eval(n.Dim, f, r, depth+1)
	}
	if !v.plain {
		// An offset of an object is read by a method of its class.
		e.implicit(n, f, r)
	}
	// An element is no object when its array holds none.
	base := v.deep
	v.Expr = n
	v.Identity = 0
	// What is known of the array says nothing of its element.
	v.NonFalse, v.NonNull = false, false
	root := n.Var
	levels := 1
	for {
		fetch, ok := root.(*syntax.ArrayDimFetch)
		if !ok {
			break
		}
		root = fetch.Var
		levels++
	}
	// What an element of the uploads table is depends on how deep it sits.
	whole := v.upload
	v.upload = 0
	v = projectValues(v)
	if v.scalars {
		// An element of such an array is one of its values.
		v.scalars = false
		v.Safe |= everyContext
	}
	if slot, constant := arrayKey(n.Dim); constant {
		if own, tracked := r.values.at(n.Var).elements.lookup(slot); tracked {
			v = resourceElement(n, own, f, r)
			v.Expr = n
			v.plain, v.deep = own.plain, own.deep
		}
	}
	if variable, ok := root.(*syntax.Variable); ok {
		key, known := literalText(n.Dim)
		if variable.NameExpr == nil && !external(variable.Name) {
			own, tracked := Value{}, false
			if slot, constant := arrayKey(n.Dim); constant && levels == 1 {
				own, tracked = f.vars.value(variable.Name).elements.lookup(slot)
			}
			switch {
			case tracked:
				// The element's own last write decides what it holds.
				v = resourceElement(n, own, f, r)
				v.Expr = n
			case whole == uploadEntry && known && uploadField(key):
				// The server chooses these fields of an uploaded file.
				v = Value{Expr: n, Complete: true}
			case whole == uploadTable:
				v.upload = uploadEntry
			case variable.Name == "GLOBALS" && levels == 1 && known && (r.global || r.globalNames[key]):
				// The entry is the variable of that name.
				held := f.vars.value(key)
				v = Value{Complete: held.Complete, Sources: held.Sources, Safe: held.Safe, Expr: n}
			}
		}
		if f.matched(requestElement(n)) {
			// A switch found the element equal to a fixed text.
			v = Value{Complete: true, Expr: n, matched: true}
		}
		if variable.Name == "_FILES" && levels == 1 {
			v.upload = uploadEntry
		}
		if variable.Name == "_SERVER" {
			if !known {
				v = Value{Expr: n}
			} else if !(strings.HasPrefix(key, "HTTP_") || key == "REQUEST_URI" || key == "QUERY_STRING" || key == "PHP_SELF" || key == "PATH_INFO" || strings.HasPrefix(key, "PHP_AUTH_")) {
				v = Value{Expr: n, Complete: true}
			}
		}
		// The fields of an entry are one level below it; deeper values
		// are what their field holds (a list of names, of sizes).
		if variable.Name == "_FILES" && levels == 2 && known {
			switch key {
			case "tmp_name", "error", "size":
				v = Value{Expr: n, Complete: true}
			case "name", "type", "full_path":
			default:
				v = Value{Expr: n}
			}
		}
	}
	v.plain, v.deep = base, base
	return v
}

// evalPropertyFetch reads a property of an object, possibly null-safe.
func (e *Env) evalPropertyFetch(n *syntax.PropertyFetch, f *frame, r *scopeResult, depth int) Value {
	e.eval(n.Var, f, r, depth+1)
	if conditional(n) {
		// The name is not evaluated when the object is null.
		read := clone(f)
		e.computed(n.Name, &read, r, depth)
		*f = rejoin(*f, read, r)
	} else {
		e.computed(n.Name, f, r, depth)
	}
	// A magic method or a hook may read the property. One that is
	// read as declared holds what its type admits.
	v := Value{}
	if r.idle() {
		return v
	}
	if !e.backed(n) {
		e.implicit(n, f, r)
		return v
	}
	property := e.types.Index.FindProperty(e.types.ClassFQN(syntax.EnclosingClass(n)), n.Name.(*syntax.Identifier).Value, e.types.PHP)
	if kind := e.kinds(property.Type, ""); kind&(kindKnown|kindObject|kindOpaque) == kindKnown {
		v.plain, v.deep = true, kind&kindArray == 0
	}
	return v
}

// uploadField names the entries of an uploaded file the server chooses.
func uploadField(key string) bool {
	return key == "tmp_name" || key == "error" || key == "size"
}
