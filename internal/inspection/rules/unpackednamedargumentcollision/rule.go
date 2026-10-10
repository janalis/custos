// Package unpackednamedargumentcollision implements the UnpackedNamedArgumentCollision inspection.
package unpackednamedargumentcollision

import (
	"math"
	"strconv"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

type rule struct{}

func New() analysis.Rule { return rule{} }
func (rule) ID() string  { return "UnpackedNamedArgumentCollision" }
func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall, syntax.KMethodCall, syntax.KStaticCall}
}

const message = "Bind each argument only once."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP81 {
		return
	}
	target, known := semanticquery.ResolveCallee(ctx, n)
	if !known {
		return
	}
	var args *syntax.ArgList
	switch call := n.(type) {
	case *syntax.FuncCall:
		args = call.Args
	case *syntax.MethodCall:
		args = call.Args
	case *syntax.StaticCall:
		args = call.Args
	}
	if args == nil {
		return
	}
	bound := map[string]bool{}
	fromSpread := map[string]bool{}
	position := 0
	for _, node := range args.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok {
			return
		}
		if arg.Unpack {
			array := semanticquery.NativeArray(ctx, arg.Value)
			if array == nil {
				return
			}
			seenKeys := map[string]bool{}
			nextIndex := int64(0)
			for _, item := range array.Items {
				key := ""
				arrayKey := ""
				integerKey := nextIndex
				if item.Key != nil {
					if text, known := semanticquery.NativeString(ctx, item.Key); known {
						if number, err := strconv.ParseInt(text, 10, 64); err == nil && strconv.FormatInt(number, 10) == text {
							integerKey = number
						} else {
							key, arrayKey = text, "s:"+text
						}
					} else if number, known := semanticquery.NativeInt(ctx, item.Key); known {
						integerKey = number
					} else {
						return
					}
				}
				if arrayKey == "" {
					// Negative implicit indices differ across PHP releases;
					// the maximum key cannot admit a following implicit item.
					if integerKey < 0 || integerKey == math.MaxInt64 {
						return
					}
					arrayKey = "i:" + strconv.FormatInt(integerKey, 10)
					if integerKey >= nextIndex {
						nextIndex = integerKey + 1
					}
				}
				// PHP replaces the value of an existing array key without
				// adding another argument or changing its insertion position.
				if seenKeys[arrayKey] {
					continue
				}
				seenKeys[arrayKey] = true
				if arrayKey[0] == 'i' {
					if position >= len(target.Params) {
						return
					}
					key = target.Params[position].Name
					position++
				}
				if bound[key] {
					ctx.ReportNode(arg, message)
				}
				bound[key] = true
				fromSpread[key] = true
			}
		} else {
			key := ""
			if arg.Name != nil {
				key = arg.Name.Value
			} else {
				if position >= len(target.Params) {
					return
				}
				key = target.Params[position].Name
				position++
			}
			if bound[key] && fromSpread[key] {
				ctx.ReportNode(arg, message)
			}
			bound[key] = true
		}
	}
}

func (rule) Semantic() {}
