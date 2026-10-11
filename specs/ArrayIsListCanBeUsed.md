---
id: ArrayIsListCanBeUsed
group: Language level migration
kind: syntax
needs: []
php: { min: "8.1", max: "" }
---

# ArrayIsListCanBeUsed

## Summary

Hand-written "is this array a list?" checks — comparing `array_values($a)`
with `$a`, or `array_keys($a)` with `range(0, count($a) - 1)` — are better
expressed with the built-in `array_is_list($a)`.

## Detection

Let `C` be a plain function call (not a method/static call) whose name, as
compared case-insensitively, is `array_values` or `array_keys` and which
resolves to that global function (unqualified or `\`-qualified; not a
qualified `Foo\array_values(...)`, a `use function` import from another
namespace, or an unqualified call in a namespace declaring that function),
with exactly one argument `A`.

- **D1** Project PHP level ≥ 8.1, where `array_is_list()` exists (custos
  diverges from the upstream threshold; see Divergences).
- **D2** `C`'s **direct** parent is a binary expression `B` (no parentheses
  in between) whose operator is the strict `===` or `!==`. Loose `==`,
  `!=` and `<>` are not reported (see Divergences). `C` may be either operand of `B`; let `O` be the other
  operand.
- **D3** (`array_values`) `O` is equivalent to `A`: same node kind and
  either structurally identical ignoring whitespace/comments, or identical
  source text. For plain variables, equal names suffice.
  `array_values($m) === $m`, `$m !== array_values($m)`.
- **D4** (`array_keys`) `O` is a plain function call to the global `range`
  (any case, resolved like `C`) with exactly two arguments where:
  - the first is the integer literal written exactly `0` (`0.0`, `00`,
    `-0`, a constant, etc. do not match);
  - the second is a binary subtraction `L - R` (not parenthesised) where
    `R` is the number literal written exactly `1` and `L` is a plain
    function call named exactly `count` with exactly one argument that is
    equivalent (as in D3) to `A`.
  `array_keys($m) === range(0, count($m) - 1)` and the mirrored
  `range(0, count($m) - 1) !== array_keys($m)` both match.
- **D5** Negation: when `B`'s operator is `!==` the replacement is negated.
- **D6** Replacement text: `("!" if negated) + Q + "array_is_list(" + <A source
  text verbatim> + ")"`, where `Q` is `\` when `C` was written fully
  qualified (`\array_values(...)`) or when an unqualified `array_is_list`
  call at that position would not reach the global function (a
  `use function` import under that name, or an `array_is_list` function
  declared in the current namespace); otherwise empty.
- **Name case.** Wherever this rule compares two expressions for
  equivalence, the names PHP resolves case-insensitively — function and
  method names, class names in calls, `new`, `instanceof` and `::`
  accesses, and keywords — compare case-insensitively (`Cache::$map['k']`
  matches `cache::$map['k']`, `$o->Name()` matches `$o->name()`); variable,
  property and constant names stay case-sensitive.

### Safety requirements

The compared array expression must be a plain variable with a fixed name.
Calls, variable variables, property reads and indexed expressions are skipped:
removing their repeated evaluation may change values or side effects.

## Exceptions (no report)

- **E1** PHP level below 8.1.
- **E2** Argument count of `C` other than 1 (`array_keys($m, 'x')`).
- **E3** `C` wrapped in parentheses, or used in a loose or non-equality
  comparison (`==`, `!=`, `<>`, `<`, `<=>`), or not inside a binary
  expression at all.
- **E4** Other operand not equivalent: `array_values($m) === $n`,
  `array_values($m) === []`.
- **E5** `range` shape mismatch: start not `0`, step argument present
  (3 arguments), end not `count(X) - 1`, `count` of something else
  (`count([])`, `count($n)`), `count` with 2 arguments, `- 2`, `+ -1`.

## Report

- Range: the whole binary expression `B` (from the start of its left operand
  to the end of its right operand).
- Severity: info (weak warning).
- Message: `Replace with '{replacement}'.`

## Fix

- **F1** Replace `B` entirely with the D6 replacement text:
  - `array_values($cfg) === $cfg` → `array_is_list($cfg)`
  - `array_keys($cfg) !== range(0, count($cfg) - 1)` → `!array_is_list($cfg)`
  - `$cfg !== \array_values($cfg)` → `!\array_is_list($cfg)`
  No parentheses are added around the replacement.

## Options

None.

## PHP versions

- Gated at ≥ 8.1, the first version providing `array_is_list()`. Upstream
  gates at ≥ 8.0; its fixture runs at 8.1.

## Examples

```php
<?php
function inspect(array $cfg, array $other) {
    return [
        <weak_warning descr="Replace with 'array_is_list($cfg)'.">array_values($cfg) === $cfg</weak_warning>,
        <weak_warning descr="Replace with '!array_is_list($cfg)'.">$cfg !== array_values($cfg)</weak_warning>,
        <weak_warning descr="Replace with '!\array_is_list($cfg)'.">\array_values($cfg) !== $cfg</weak_warning>,
        <weak_warning descr="Replace with '$cfg !== [] &amp;&amp; array_is_list($cfg)'.">array_keys($cfg) === range(0, count($cfg) - 1)</weak_warning>,
        <weak_warning descr="Replace with '$cfg === [] || !array_is_list($cfg)'.">range(0, count($cfg) - 1) !== array_keys($cfg)</weak_warning>,
        array_values($this->rows['a']) === $this->rows[ 'a' ],
    ];
}

function wrapped(array $cfg): bool
{
    return $cfg && !(<weak_warning descr="Replace with '$cfg !== [] &amp;&amp; array_is_list($cfg)'.">array_keys($cfg) === range(0, count($cfg) - 1)</weak_warning>)
        || <weak_warning descr="Replace with '$cfg === [] || !array_is_list($cfg)'.">array_keys($cfg) !== range(0, count($cfg) - 1)</weak_warning> === false;
}
```

```php
<?php
function inspect(array $cfg, array $other) {
    return [
        array_is_list($cfg),
        !array_is_list($cfg),
        !\array_is_list($cfg),
        $cfg !== [] && array_is_list($cfg),
        $cfg === [] || !array_is_list($cfg),
        array_values($this->rows['a']) === $this->rows[ 'a' ],
    ];
}

function wrapped(array $cfg): bool
{
    return $cfg && !($cfg !== [] && array_is_list($cfg))
        || ($cfg === [] || !array_is_list($cfg)) === false;
}
```

## Divergences

- **Audit safety (custos diverges).** Only fixed-name variables qualify. Structural equivalence of calls or
  property/index reads does not establish stable repeated evaluation.

- Version threshold (custos diverges): upstream enables the rule from
  PHP 8.0, but `array_is_list()` is a PHP 8.1 function, so the fix would
  break 8.0 code. custos gates at ≥ 8.1 (the upstream fixture runs at 8.1,
  so conformance is unaffected).
- Inserting `!array_is_list(...)` where `B` was an operand of a
  higher-precedence operator is safe (function call binds tighter); no
  precedence issue expected.
- custos diverges from upstream on loose comparisons (D2, E3). Upstream
  also rewrites `==`/`!=`, but loose array equality compares key/value
  pairs regardless of order and with loose value comparison:
  `array_values([1 => 'a', 0 => 'a']) == [1 => 'a', 0 => 'a']` is true
  although the array is not a list, and a key such as `'01'` loosely equals
  `1` in the `range` form. The rewrite would change the result, so custos
  reports only the strict `===`/`!==` forms. The affected EA case is listed
  in `testdata/ea-divergences.json`.
- **Function names (custos diverges):** upstream matches `array_values`,
  `array_keys`, `range` and `count` on the written last segment,
  case-sensitively: it misses `Array_Values($a) === $a` and reports user
  functions with those names (e.g. a namespaced `range()`). custos compares
  the names case-insensitively and requires each call to resolve to the
  global function (C, D4, and the inner `count` call).
- **Builtin spelling (custos diverges).** Upstream copies the qualifier
  written before `array_values`/`array_keys`, so an unqualified call in a
  namespace that declares or imports its own `array_is_list` is rewritten
  to call that function. custos writes `\array_is_list(` in that case (D6).
- **Name case (custos diverges).** Upstream compares the expressions
  textually, so operands that differ only in the case of a function, method
  or class name (`Stats::$n` vs `stats::$n`), which PHP treats as the same,
  are not recognised as equivalent. custos folds the case of those names
  (Detection, "Name case").
- **Empty array in the `range()` form (custos diverges).**
  `array_keys($a) === range(0, count($a) - 1)` is false for `[]`
  (`range(0, -1)` is `[0, -1]`) while `array_is_list([])` is true. custos
  keeps the meaning: the replacement (message and fix) is
  `$a !== [] && array_is_list($a)`, negated `$a === [] || !array_is_list($a)`,
  parenthesized when the comparison is an operand of another operator.
  The `array_values()` form needs no such guard.
- **`NAN` in a hash-backed list (accepted, fix kept).** The
  `array_values()` form is not strictly identical to `array_is_list()` in
  one corner: when the array is a list by its keys but is not stored packed
  (for instance `[1 => 1, 0 => NAN]` after `unset($a[1])`), `array_values()`
  builds a copy, the strict comparison then compares the elements one by
  one, and `NAN !== NAN` makes the hand-written check false although
  `array_is_list()` answers true. A packed `[NAN]` is returned as is and
  compares true, so only such re-packed lists holding `NAN` differ. The
  hand-written result is the accidental one there (the array is a list),
  and withholding the fix whenever an element might be a float would drop
  nearly every useful fix, so the report and the fix are kept.
