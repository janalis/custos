// Package uploadclientmimetrusted implements the native UploadClientMimeTrusted inspection.
package uploadclientmimetrusted

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Inspect uploaded content instead of trusting its client MIME label."

// contentInspectors read the stored bytes of the file they are given.
var contentInspectors = map[string]bool{
	"finfo_file":        true,
	"mime_content_type": true,
	"getimagesize":      true,
	"exif_imagetype":    true,
}

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "UploadClientMimeTrusted" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "move_uploaded_file") {
		return
	}
	upload, known := uploadEntry(ctx, semanticquery.CallArgument(call.Args, 0, "from"), "tmp_name")
	if !known || !trustsClientType(ctx, call, upload) || contentInspected(ctx, call, upload) {
		return
	}
	ctx.ReportNode(call, message)
}

// uploadEntry returns the literal $_FILES key of `$_FILES[key][field]`.
func uploadEntry(ctx *analysis.Context, expr syntax.Expr, field string) (string, bool) {
	fetch, ok := expr.(*syntax.ArrayDimFetch)
	if !ok {
		return "", false
	}
	key, known := semanticquery.NativeString(ctx, fetch.Dim)
	if !known || key != field {
		return "", false
	}
	entry, ok := fetch.Var.(*syntax.ArrayDimFetch)
	if !ok {
		return "", false
	}
	root, ok := entry.Var.(*syntax.Variable)
	if !ok || root.Name != "_FILES" {
		return "", false
	}
	return semanticquery.NativeString(ctx, entry.Dim)
}

// trustsClientType reports whether the call is reached only when the client
// type of the same upload matched an allowed MIME string: it sits in the
// accepting branch of such a test, in an alternative taken once the mismatch
// test failed, or after a guard clause leaving on mismatch.
func trustsClientType(ctx *analysis.Context, call *syntax.FuncCall, upload string) bool {
	for node := syntax.Node(call); node != nil && !syntax.IsVariableScope(node); node = node.Parent() {
		switch branch := node.Parent().(type) {
		case *syntax.If:
			if node == branch.Body && clientTypeTest(ctx, branch.Cond, upload, true) {
				return true
			}
			switch node.(type) {
			case *syntax.ElseIf, *syntax.Else:
				if failedEarlier(ctx, branch, node, upload) {
					return true
				}
			}
		case *syntax.ElseIf:
			if clientTypeTest(ctx, branch.Cond, upload, true) {
				return true
			}
		}
		stmt, ok := node.(syntax.Stmt)
		if !ok {
			continue
		}
		list, index, ok := astquery.StmtList(ctx.File, stmt)
		if !ok {
			continue
		}
		for _, earlier := range list[:index] {
			guard, ok := earlier.(*syntax.If)
			if ok && syntax.Terminates(guard.Body) && clientTypeTest(ctx, guard.Cond, upload, false) {
				return true
			}
		}
	}
	return false
}

// failedEarlier reports whether the alternative alt of chain (an elseif or
// the else) is reached only after a condition that is true whenever the
// client type test fails: the `if` condition or an earlier `elseif` one.
func failedEarlier(ctx *analysis.Context, chain *syntax.If, alt syntax.Node, upload string) bool {
	if clientTypeTest(ctx, chain.Cond, upload, false) {
		return true
	}
	for _, earlier := range chain.ElseIfs {
		if syntax.Node(earlier) == alt {
			break
		}
		if clientTypeTest(ctx, earlier.Cond, upload, false) {
			return true
		}
	}
	return false
}

// clientTypeTest reports whether cond being true (accept) or false (!accept)
// implies that the upload's client type matched an allowed MIME string.
func clientTypeTest(ctx *analysis.Context, cond syntax.Expr, upload string, accept bool) bool {
	switch e := syntax.UnwrapParens(cond).(type) {
	case *syntax.Unary:
		return e.Op.Kind == syntax.TExclaim && clientTypeTest(ctx, e.Expr, upload, !accept)
	case *syntax.Binary:
		switch e.Op.Kind {
		case syntax.TBooleanAnd, syntax.TBooleanOr:
			// Every `&&` operand holds in an accepting branch; every `||`
			// operand is false once a guard clause was passed.
			return (e.Op.Kind == syntax.TBooleanAnd) == accept &&
				(clientTypeTest(ctx, e.Left, upload, accept) || clientTypeTest(ctx, e.Right, upload, accept))
		case syntax.TIsIdentical, syntax.TIsEqual, syntax.TIsNotIdentical, syntax.TIsNotEqual:
			positive := e.Op.Kind == syntax.TIsIdentical || e.Op.Kind == syntax.TIsEqual
			return positive == accept &&
				(clientTypeIs(ctx, e.Left, e.Right, upload) || clientTypeIs(ctx, e.Right, e.Left, upload))
		}
	case *syntax.FuncCall:
		return accept && semanticquery.NativeBuiltin(ctx, e, "in_array") &&
			clientType(ctx, semanticquery.CallArgument(e.Args, 0, "needle"), upload) &&
			mimeList(ctx, semanticquery.CallArgument(e.Args, 1, "haystack"))
	}
	return false
}

func clientType(ctx *analysis.Context, expr syntax.Expr, upload string) bool {
	entry, known := uploadEntry(ctx, expr, "type")
	return known && entry == upload
}

func clientTypeIs(ctx *analysis.Context, subject, value syntax.Expr, upload string) bool {
	return clientType(ctx, subject, upload) && mimeString(ctx, value)
}

func mimeString(ctx *analysis.Context, expr syntax.Expr) bool {
	mime, known := semanticquery.NativeString(ctx, expr)
	return known && strings.Contains(mime, "/")
}

// mimeList reports whether expr is a non-empty array of MIME string literals.
func mimeList(ctx *analysis.Context, expr syntax.Expr) bool {
	list := semanticquery.NativeArray(ctx, expr)
	if list == nil || len(list.Items) == 0 {
		return false
	}
	for _, item := range list.Items {
		if !mimeString(ctx, item.Value) {
			return false
		}
	}
	return true
}

// contentInspected reports whether the same uploaded temporary file is handed
// to a content inspector (fileinfo, image probing or an image decoder)
// anywhere before the call in its scope. Dominance is not required: an
// acceptance path that may depend on stored bytes is not provably label-only.
func contentInspected(ctx *analysis.Context, call *syntax.FuncCall, upload string) bool {
	aliases := map[string]bool{}
	found := false
	visit := func(n syntax.Node) bool {
		if found || n.Span().Start >= call.Span().Start {
			return false
		}
		var args *syntax.ArgList
		switch e := n.(type) {
		case *syntax.Assign:
			target, ok := e.Var.(*syntax.Variable)
			if entry, known := uploadEntry(ctx, e.Value, "tmp_name"); ok && known && entry == upload {
				aliases[target.Name] = true
			}
		case *syntax.FuncCall:
			name := semanticquery.NativeBuiltinName(ctx, e)
			if contentInspectors[name] || strings.HasPrefix(name, "imagecreatefrom") {
				args = e.Args
			}
		case *syntax.MethodCall:
			// finfo::file(); the receiver type is not required to be proven.
			if name, ok := e.Name.(*syntax.Identifier); ok && strings.EqualFold(name.Value, "file") {
				args = e.Args
			}
		}
		for _, node := range astquery.ArgumentNodes(args) {
			arg, ok := node.(*syntax.Arg)
			if !ok {
				continue
			}
			if entry, known := uploadEntry(ctx, arg.Value, "tmp_name"); known && entry == upload {
				found = true
			}
			if local, ok := arg.Value.(*syntax.Variable); ok && aliases[local.Name] {
				found = true
			}
		}
		return true
	}
	if scope := syntax.EnclosingVariableScope(call); scope != nil {
		syntax.Inspect(scope, visit)
	} else {
		syntax.InspectFile(ctx.File, visit)
	}
	return found
}
