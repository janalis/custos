// Package uploadclientmimetrusted implements the native UploadClientMimeTrusted inspection.
package uploadclientmimetrusted

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
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
// type of the same upload matched an allowed MIME string: where it runs, the
// conditions guarding it narrow `$_FILES[key]['type']` to a list of such
// strings.
func trustsClientType(ctx *analysis.Context, call *syntax.FuncCall, upload string) bool {
	if semanticquery.NativeLabelled(ctx, call) {
		return false // a jump may land between the test and the call
	}
	var label syntax.Expr
	visit := func(n syntax.Node) bool {
		if label != nil || n.Span().Start >= call.Span().Start || syntax.IsVariableScope(n) {
			return false
		}
		if x, ok := n.(syntax.Expr); ok && clientType(ctx, x, upload) {
			label = x
		}
		return true
	}
	if scope := syntax.EnclosingVariableScope(call); scope != nil {
		syntax.Inspect(syntax.VariableScopeBody(scope), visit)
	} else {
		syntax.InspectFile(ctx.File, visit)
	}
	if label == nil {
		return false
	}
	values, ok := ctx.Types().Native().NarrowedMemberAt(label, call, types.TrackedStrings()).StringValueList()
	if !ok {
		return false
	}
	for _, v := range values {
		if !strings.Contains(v, "/") {
			return false
		}
	}
	return true
}

func clientType(ctx *analysis.Context, expr syntax.Expr, upload string) bool {
	entry, known := uploadEntry(ctx, expr, "type")
	return known && entry == upload
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
