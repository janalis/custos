package phpunit

import (
	"slices"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// messagePosition gives, for PHPUnit's common assertions, the position of
// their trailing `$message` parameter, used when PHPUnit is not indexed.
var messagePosition = map[string]int{
	"asserttrue": 1, "assertfalse": 1, "assertnull": 1, "assertnotnull": 1,
	"assertempty": 1, "assertnotempty": 1, "assertnottrue": 1, "assertnotfalse": 1,
	"assertisarray": 1, "assertisbool": 1, "assertisfloat": 1, "assertisint": 1,
	"assertisnumeric": 1, "assertisobject": 1, "assertisresource": 1, "assertisstring": 1,
	"assertisscalar": 1, "assertiscallable": 1, "assertisiterable": 1,
	"assertfileexists": 1, "assertfiledoesnotexist": 1, "assertdirectoryexists": 1,
	"assertjson": 1, "assertnan": 1, "assertfinite": 1, "assertinfinite": 1,
	"assertsame": 2, "assertnotsame": 2, "assertequals": 2, "assertnotequals": 2,
	"assertequalscanonicalizing": 2, "assertequalsignoringcase": 2,
	"assertcount": 2, "assertnotcount": 2, "assertcontains": 2, "assertnotcontains": 2,
	"assertstringcontainsstring": 2, "assertstringnotcontainsstring": 2,
	"assertstringcontainsstringignoringcase": 2, "assertstringstartswith": 2,
	"assertstringendswith": 2, "assertstringstartsnotwith": 2, "assertstringendsnotwith": 2,
	"assertinstanceof": 2, "assertnotinstanceof": 2, "assertarrayhaskey": 2,
	"assertarraynothaskey": 2, "assertmatchesregularexpression": 2,
	"assertdoesnotmatchregularexpression": 2, "assertgreaterthan": 2,
	"assertgreaterthanorequal": 2, "assertlessthan": 2, "assertlessthanorequal": 2,
	"assertjsonstringequalsjsonstring": 2, "assertstringequalsfile": 2,
	"assertfileequals": 2, "assertsamesize": 2, "assertthat": 2,
	"assertequalswithdelta": 3,
}

// MessageArgument reports whether arg is the `$message` argument of PHPUnit
// assertion call (one PuAssertion or UnindexedAssertion accepts): named so, or at the position of the parameter of that
// name, as indexed or, for a common assertion of an unindexed PHPUnit, as
// PHPUnit declares it. The message only describes a failure.
func MessageArgument(ctx *analysis.Context, call syntax.Node, arg *syntax.Arg) bool {
	if arg.Name != nil {
		return strings.EqualFold(arg.Name.Value, "message")
	}
	// Positional arguments come first: an argument's index is its position.
	var list *syntax.ArgList
	var name string
	var params []index.Param
	indexed := false
	if f, ok := call.(*syntax.FuncCall); ok {
		list, params, indexed = f.Args, ctx.Types().ResolveFunction(f).Params, true
	} else {
		c, _ := AsPuCall(call)
		list, name = c.Args, c.Ident.Value
		params, indexed = assertionParams(ctx, c)
	}
	at := slices.Index(list.Args, syntax.Expr(arg))
	if indexed {
		return at < len(params) && params[at].Name == "message"
	}
	position, known := messagePosition[strings.ToLower(name)]
	return known && at == position
}

// assertionParams returns the indexed parameters of the assertion c calls.
func assertionParams(ctx *analysis.Context, c PuCall) ([]index.Param, bool) {
	var classes []string
	if n, named := c.Recv.(*syntax.Name); named {
		classes = []string{PuResolveClassName(ctx, n)}
	} else {
		classes = ctx.TypeOf(c.Recv).Classes()
	}
	for _, class := range classes {
		if m := ctx.Index().FindMethod(strings.TrimPrefix(class, `\`), c.Ident.Value, ctx.PHP); m != nil {
			return m.Params, true
		}
	}
	return nil, false
}
