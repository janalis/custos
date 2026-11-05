package infer

import (
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

func sliceArrayCount(result, input types.Type, offset, length syntax.Expr) types.Type {
	low, high, known := input.ArrayCardinality()
	start, fixed := integerLiteral(offset)
	limit, bounded := integerLiteral(length)
	omitted := length == nil || constLiteral(length) == "null"
	if !known || !fixed || !bounded && !omitted {
		return preserveArrayCount(result, input, true)
	}
	count := func(n int64) int {
		at := min(n, start)
		if start < 0 {
			at = max(int64(0), n+start)
		}
		remaining := n - at
		if omitted {
			return int(remaining)
		}
		if limit >= 0 {
			return int(min(remaining, limit))
		}
		return int(max(at, n+limit) - at)
	}
	return result.WithArrayCardinality(count(int64(low)), count(int64(high)))
}
