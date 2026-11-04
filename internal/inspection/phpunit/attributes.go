package phpunit

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// AttributeMetadata normalizes recognized PHPUnit attributes to their
// annotation equivalents. Names resolve through imports; unrelated short
// attribute names and uncertain argument bindings provide no metadata.
func AttributeMetadata(ctx *analysis.Context, groups []*syntax.AttributeGroup) []Metadata {
	if ctx.PHP < phpversion.PHP80 {
		return nil
	}
	var out []Metadata
	for _, group := range groups {
		for _, attribute := range group.Attrs {
			name := ctx.Names().Class(attribute.Name.Value, attribute.Name.Span().Start)
			var tag, parameter string
			switch strings.ToLower(strings.TrimPrefix(name, `\`)) {
			case `phpunit\framework\attributes\dataprovider`:
				tag, parameter = "@dataProvider", "methodName"
			case `phpunit\framework\attributes\depends`:
				tag, parameter = "@depends", "methodName"
			case `phpunit\framework\attributes\coversclass`:
				tag, parameter = "@covers", "className"
			case `phpunit\framework\attributes\test`:
				tag = "@test"
			default:
				continue
			}
			metadata := Metadata{Name: tag, Start: attribute.Name.Span().Start, End: attribute.Name.Span().End, Annotation: true, Attribute: true}
			if tag == "@test" {
				out = append(out, metadata)
				continue
			}
			value := metadataArgument(attribute.Args, parameter)
			if value == nil {
				continue
			}
			if literal, ok := syntax.UnwrapParens(value).(*syntax.Literal); ok {
				text, known := syntax.LiteralStringValue(literal)
				if !known {
					continue
				}
				metadata.Value = text
				if tag == "@covers" {
					metadata.Value = `\` + strings.TrimPrefix(text, `\`)
				}
			} else if class, ok := syntax.UnwrapParens(value).(*syntax.ClassConstFetch); ok && tag == "@covers" {
				member, memberOK := class.Name.(*syntax.Identifier)
				name, nameOK := class.Class.(*syntax.Name)
				if !memberOK || !nameOK || !strings.EqualFold(member.Value, "class") {
					continue
				}
				resolved := ctx.Types().ClassRef(name)
				if resolved == "" {
					continue
				}
				metadata.Value = `\` + strings.TrimPrefix(resolved, `\`)
			} else {
				continue
			}
			out = append(out, metadata)
		}
	}
	return out
}

func metadataArgument(list *syntax.ArgList, parameter string) syntax.Expr {
	if list == nil {
		return nil
	}
	var value syntax.Expr
	named := false
	seen := map[string]bool{}
	for i, node := range list.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok || arg.Unpack {
			return nil
		}
		match := i == 0 && arg.Name == nil
		if arg.Name != nil {
			if seen[arg.Name.Value] {
				return nil
			}
			seen[arg.Name.Value], named = true, true
			match = arg.Name.Value == parameter
		} else if named {
			return nil
		}
		if match {
			if value != nil {
				return nil
			}
			value = arg.Value
		}
	}
	return value
}
