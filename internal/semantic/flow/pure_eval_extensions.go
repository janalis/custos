package flow

import (
	"strconv"
	"strings"

	"custos/internal/php/syntax"
)

func (e *Env) evalPureExtended(expr *pureExpr, slots []Value, r *scopeResult, depth int) (Value, bool) {
	if expr.kind == pureInline {
		return Value{}, false
	}
	if expr.kind == pureFold {
		return e.evalPureFold(expr, slots, r, depth)
	}
	if expr.kind == pureFind {
		return e.evalPureFind(expr, slots, r, depth)
	}
	if expr.kind == pureDestructure {
		return e.evalDestructure(expr, slots, r, depth)
	}
	if expr.kind == pureLet {
		return e.evalPureLet(expr, slots, r, depth)
	}
	if expr.kind == pureSpreadArray {
		return e.evalSpreadArray(expr, slots, r, depth)
	}
	if expr.kind == pureReduce || expr.kind == pureBuild || expr.kind == pureFilterBuild || expr.kind == pureGuardReduce {
		return e.evalPureReduction(expr, slots, r, depth)
	}
	if expr.kind == pureSwitch {
		return e.evalIntegerSwitch(expr, slots, r, depth)
	}
	if expr.kind == pureMatch {
		return e.evalPureMatch(expr, slots, r, depth)
	}
	if expr.kind == pureArray {
		set := &elementSet{whole: true, items: map[string]Value{}, width: len(expr.args) / 2}
		value := Value{Complete: true, plain: true, deep: true}
		r.copies += len(expr.args) / 2
		if r.copies > MaxCopied {
			return Value{}, false
		}
		for i := 0; i < len(expr.args); i += 2 {
			entry, ok := e.evalPure(expr.args[i+1], slots, r, depth-1)
			if !ok || !entry.Complete || !entry.plain || !entry.deep || entry.Identity != 0 || entry.storedIdentity != 0 || entry.elements != nil && entry.elements.hasResources {
				return Value{}, false
			}
			key := expr.args[i].literal.s
			set.put(key, entry)
			set.order = append(set.order, key)
			value = merge(value, entry)
		}
		value.elements = set
		value.plain, value.deep = true, true
		return value, true
	}
	left, ok := e.evalPure(expr.left, slots, r, depth-1)
	if !ok {
		return Value{}, false
	}
	if expr.kind >= pureFill && expr.kind <= pureSum {
		return e.evalArrayConstruction(expr, left, slots, r, depth)
	}
	if expr.kind == pureInteger {
		return e.evalPureInteger(expr, left, slots, r, depth)
	}
	if expr.kind >= pureCase && expr.kind <= pureStrPad {
		return e.evalTextShape(expr, left, slots, r, depth)
	}
	if expr.kind == pureArrayPad || expr.kind == pureRange {
		return e.evalPadding(expr, left, slots, r, depth)
	}
	if expr.kind >= pureUnique && expr.kind <= pureTextScan {
		return e.evalQuery(expr, left, slots, r, depth)
	}
	if expr.kind == pureReplace || expr.kind == pureRepeat {
		return e.evalByteTransform(expr, left, slots, r, depth)
	}
	if expr.kind == pureReverse {
		return e.evalPureReverse(expr, left, slots, r, depth)
	}
	if expr.kind == pureJoin || expr.kind == pureSplit {
		return e.evalStringTransform(expr, left, slots, r, depth)
	}
	if expr.kind == pureColumn || expr.kind == pureMerge || expr.kind == pureCombine || expr.kind == pureArrayReplace {
		return e.evalArrayTransform(expr, left, slots, r, depth)
	}
	if expr.kind == pureSlice {
		return e.evalPureSlice(expr, left, slots, r, depth)
	}
	if expr.kind == pureFilteredKeys {
		return e.evalFilteredKeys(expr, left, slots, r, depth)
	}
	if expr.kind == pureCast {
		return evalPureCast(expr, left)
	}
	if expr.kind == pureHigher {
		return e.evalPureHigher(expr, left, slots, r, depth)
	}
	if expr.kind == pureSequence {
		return e.evalPure(expr.right, slots, r, depth-1)
	}
	if expr.kind == pureIndex || expr.kind == pureQuietIndex || expr.kind == pureExists {
		// PHP evaluates offset expressions even on a quiet missing path.
		keyValue, known := e.evalPure(expr.right, slots, r, depth-1)
		key, scalar := knownScalar(keyValue)
		if !known || !scalar || key.k != 'i' && key.k != 's' {
			return Value{}, false
		}
		if scalar, known := knownScalar(left); expr.kind == pureQuietIndex && known && scalar.k == 'n' {
			return left, true
		}
		if !left.Complete || !left.plain || !left.deep || left.elements == nil || !left.elements.whole || left.elements.optional || left.elements.hasResources {
			return Value{}, false
		}

		text := key.s
		if key.k == 'i' {
			text = strconv.FormatInt(key.i, 10)
		}
		value, held := left.elements.lookup(text)
		if expr.kind == pureExists {
			return scalarValue(constantScalar{k: 'b', b: held}), true
		}
		if !held && expr.kind == pureQuietIndex {
			return scalarValue(constantScalar{k: 'n'}), true
		}
		return value, held && value.Complete && !value.replaced && value.Invalidated == 0
	}
	if expr.kind == pureNonNull || expr.kind == pureShort && expr.op == syntax.TCoalesce {
		nonnull, known := pureNonNullValue(left)
		if !known {
			return Value{}, false
		}
		if expr.kind == pureNonNull {
			return scalarValue(constantScalar{k: 'b', b: nonnull}), true
		}
		if nonnull {
			return left, true
		}
		return e.evalPure(expr.right, slots, r, depth-1)
	}
	if expr.kind == pureOrderedArray || expr.kind == pureTypeTest || expr.kind == pureSubstring {
		return e.evalCollectionNative(expr, left, slots, r, depth)
	}
	if expr.kind == pureStringPredicate {
		return e.evalStringPredicate(expr, left, slots, r, depth)
	}
	if expr.kind == pureCount || expr.kind == pureMember {
		return e.evalFiniteNative(expr, left, slots, r, depth)
	}
	if expr.kind == pureChoice || expr.kind == pureShort || expr.kind == pureEmpty {
		return e.evalPureTruth(expr, left, slots, r, depth)
	}
	scalar, known := knownScalar(left)
	if !known {
		return Value{}, false
	}
	switch expr.kind {
	case pureLength, pureTrim:
		if scalar.k != 's' || len(scalar.s) > 4096 {
			return Value{}, false
		}
		if expr.kind == pureLength {
			value := scalarValue(constantScalar{k: 'i', i: int64(len(scalar.s))})
			value.Sources = left.Sources
			value.Safe = everyContext
			return value, true
		}
		return pureSliceValue(left, strings.Trim(scalar.s, " \n\r\t\v\x00")), true
	}
	return Value{}, false
}
