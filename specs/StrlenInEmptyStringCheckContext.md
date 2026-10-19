---
id: StrlenInEmptyStringCheckContext
group: Control flow
kind: semantic
needs: [types, stubs]
php: { min: "", max: "" }
---

# StrlenInEmptyStringCheckContext

## Summary

Measuring a string's length only to know whether it is empty
(`strlen($s) > 0`, `!strlen($s)`, `if (strlen($s))`) is indirect and
slower than comparing with the empty string. Suggest an identity comparison
with `''` instead, casting to string when the value is not known to be a
string.

## Detection

- **D1** A plain function call that resolves to the global function
  `strlen` — name compared case-insensitively as PHP does
  (`StrLen`, `\STRLEN` match); written with a single leading `\`, or
  unqualified with no function of that name declared or imported in the
  current namespace (`Ns\strlen(...)` and shadowing user functions are not
  matched) — with exactly one argument. Call it
  `L`, its first argument `A`.
- **D2** `L` may be located anywhere: inside a function-like or class body,
  or in top-level file code (custos diverges, see Divergences).

Then the first matching pattern below applies (each yields a *target* node
and an *empty* flag: empty = the expression is true when the string is
empty).

Numeric comparisons — `L`'s **direct** parent (no parentheses in between) is
a binary expression whose other operand is a number literal (integer/float
token, or unary minus applied to one). The number is compared by its exact
source text:

- **D3** operator `>`, `L` is the left operand, number text `0`
  → target = binary, empty = false (`strlen($s) > 0`).
- **D4** operator `<` or `>=`, `L` is the left operand, number text `1`
  → target = binary; empty = true for `<`, false for `>=`.
- **D5** otherwise, operator is one of `==`, `!=` (also `<>`), `===`, `!==`,
  number text `0`, `L` on **either** side → target = binary; empty = true for
  `==`/`===`, false for `!=`/`<>`/`!==`.

Boolean contexts — when D3–D5 did not match:

- **D6** `L`, after skipping any wrapping parentheses, is used as a logical
  operand, i.e. the first non-parenthesis ancestor is:
  - the condition of an `if`, `elseif`, `while` or `do … while`;
  - a logical-not `!` expression;
  - either operand of `&&`, `||`, `and`, `or`;
  - the condition (first operand) of a full ternary `c ? a : b` (not the
    short `?:` form, and not a branch).
- **D7** For D6, target and flag:
  - if `L`'s **direct** parent (no parentheses in between) is a `!`
    expression, target = that `!` expression, empty = true
    (`!strlen($s)`);
  - otherwise target = `L` itself, empty = false. This includes
    `!(strlen($s))` (parenthesised): the target is just `L`, the `!` and
    parentheses stay.

*Cast decision*: `A` is considered a known string when its inferred type set
(declarations, docblocks, assignments, function return types from stubs —
e.g. `trim()` returns `string`) contains **exactly one** type and that type
is `string`. Anything else — unknown type, `?string` / `string|null`,
`int`, `float`, unions — needs a cast.

### Safety requirements

Only the byte-based built-in `strlen` with exactly one plain argument
qualifies. Multibyte length calls are not reported, regardless of whether
an encoding argument is explicit.

## Exceptions (no report)

- **E1** (none — top-level code is reported too, see D2.)
- **E2** Yoda-ordered threshold comparisons `1 > strlen($s)`,
  `1 <= strlen($s)`, `0 < strlen($s)`: only the forms in D3/D4 with `L` on
  the left are recognised.
- **E3** Other thresholds or number spellings: `strlen($s) > 1`,
  `strlen($s) == 1`, `strlen($s) >= 0`, `strlen($s) === 0.0`,
  `strlen($s) == '0'` (string, not a number literal), comparisons with a
  constant or variable.
- **E4** Parenthesised call as a comparison operand
  (`(strlen($s)) > 0`): the direct parent is not the binary expression.
- **E5** Non-boolean uses: `return strlen($s);`, `$n = strlen($s);`,
  arithmetic, `for (; strlen($s); )` conditions, `xor`, short-ternary
  `strlen($s) ?: 1`, ternary branches, casts (`(bool) strlen($s)`).
- **E6** Method or static calls named `strlen` / `mb_strlen`; calls without
  arguments.

## Report

- Range: the target node — the whole binary expression (D3–D5, from the
  start of the left operand to the end of the right operand, inner spacing
  included), the whole `!…` expression (D7 first case), or the call `L`
  (D7 second case, from the function name through `)`).
- Severity: info (weak warning).
- Message: `Compare with an empty string instead: '{replacement}'.` where
  `{replacement}` is the F1 text.

## Fix

- **F1** Replace the target with a comparison:
  - operator `{op}` = `===` when empty is true, `!==` otherwise;
  - operand `{v}` = `A`'s source text verbatim, prefixed with `(string)`
    (**no space** after the cast — upstream's formatter normalises it, and
    the fix output is compared whitespace-collapsed) unless `A` is a known
    string (see cast decision).
  - **F1a** Parenthesise `A` (custos diverges, see Divergences): with the
    cast, wrap `A` in `(…)` when it is a binary operation, `instanceof`,
    ternary (full or short), assignment, or a low-precedence keyword
    expression (`print`, `yield`, `include`/`require`, `throw`, arrow
    function) — `(string)($raw ?? $alt)`; without the cast, wrap it when it
    binds as loosely as or more loosely than `===`: logical (`&&`, `||`,
    `and`, `or`, `xor`), bitwise (`&`, `|`, `^`), `??`, equality operators
    (`==`, `!=`, `===`, `!==`, `<=>`), ternaries, assignments and the same
    keyword expressions — `($p ?: $q) !== ''`. Tighter-binding operands
    such as `$p . $tail` are left as is.
  - **F1b** Parenthesise the whole replacement (custos diverges, see
    Divergences) when the target is a direct operand of something binding
    at least as tightly as `===`: a unary operator or cast (`!`, `-`, `+`,
    `~`, `@`, `(int)`…), `instanceof`, or a binary operator other than the
    logical (`&&`, `||`, `and`, `or`, `xor`), bitwise (`&`, `|`, `^`) and
    `??` ones — so comparison, equality, `<=>`, `|>`, `.`, shift and
    arithmetic parents all get `(…)`: `!!strlen($s)` → `!((string)$s === '')`,
    `strlen($s) > 0 === $on` → `((string)$s !== '') === $on`. Both operand
    orders are wrapped alike. The reported message keeps the bare
    comparison text.
  - The operand order follows the global comparison style setting
    (`comparisonStyle`, default `regular`):
    - regular: `{v} {op} ''`
    - yoda: `'' {op} {v}`
  Single spaces around `{op}`; `''` is two single quotes. The upstream
  fixture runs in **yoda** style, so conformance expects `'' === $s`,
  `'' !== (string)$n`.
- Examples (regular): `strlen($name) > 0` → `$name !== ''`;
  `!strlen($id)` with unknown `$id` → `(string)$id === ''`.
  Examples (yoda): the same give `'' !== $name` and `'' === (string)$id`.

## Options

| Option | Type | Default | Effect |
|--------|------|---------|--------|
| (none) | | | Operand order follows the global comparison style (`regular` default / `yoda`). |

## PHP versions

No gating. The upstream fixture has no explicit language level (IDE test
default, below 7.1) yet uses nullable parameter types (`?string`); the parser
must accept them and the type inference must treat `?T` as `T|null`
regardless of the configured version.

## Examples

```php
<?php
class Form
{
    public function check(string $name, $id, int $n)
    {
        $a = <weak_warning descr="Compare with an empty string instead: '$name !== '''.">strlen($name) > 0</weak_warning>;
        $b = !mb_strlen($id);
        $c = <weak_warning descr="Compare with an empty string instead: '(string)$n !== '''.">strlen($n) != 0</weak_warning>;
        $d = $a || <weak_warning descr="Compare with an empty string instead: '$name !== '''.">\strlen($name)</weak_warning>;
        do {} while (<weak_warning descr="Compare with an empty string instead: '$name === '''.">strlen($name) == 0</weak_warning>);
        $f = fn($s) => <weak_warning descr="Compare with an empty string instead: '(string)$s !== '''.">strlen($s) !== 0</weak_warning>;
        return [$a, $b, $c, $d, $f];
    }
}
```

```php
<?php
class Form
{
    public function check(string $name, $id, int $n)
    {
        $a = $name !== '';
        $b = !mb_strlen($id);
        $c = (string)$n !== '';
        $d = $a || $name !== '';
        do {} while ($name === '');
        $f = fn($s) => (string)$s !== '';
        return [$a, $b, $c, $d, $f];
    }
}
```

## Divergences

- **Audit safety (custos diverges).** Multibyte length checks are skipped: some encodings represent zero
  characters with nonempty bytes, and removing an encoding argument can
  discard side effects. Extra strlen arguments are also not discarded.

- **Parenthesised argument (custos diverges).** Upstream inserts `A`'s
  text unparenthesised, so low-precedence arguments produce wrong code:
  `strlen($p ?: $q) > 0` → `'' !== $p ?: $q`; with a cast,
  `(string)$p . $q` / `(string)$p + $q` binds the cast to `$p` only.
  custos wraps `A` in parentheses when needed (F1a): `($p ?: $q) !== ''`,
  `(string)($raw + 1) !== ''`. Not covered by upstream fixtures.
- Replacing a bare `L` inside a `&&`/`||`/`and`/`or` operand keeps
  precedence correct for `===`/`!==` in all those contexts, so no change is
  needed there.
- **Parenthesised replacement (custos diverges).** A target such as
  `!strlen($s)` or `strlen($s) > 0` can itself be the operand of a
  tighter-binding or non-associative operator. Dropping the comparison in
  unparenthesised then regroups the expression (`!!strlen($s)` would become
  `!(string)$s === ''`, negating the string first; `!strlen($s) . 'x'`
  would compare with `'' . 'x'`) or chains two comparisons, which PHP
  rejects (`!strlen($a) === !strlen($b)`). custos wraps the replacement in
  parentheses in those positions (F1b). Not covered by upstream fixtures.
- **Top-level code (custos diverges).** Upstream only reports calls that
  have a function-like or class ancestor, so the same `strlen($s) > 0` in a
  plain script file goes unreported. Nothing about the pattern depends on
  the enclosing scope, so custos reports it everywhere (D2).
- **Function name (custos diverges).** Upstream matches the bare name
  case-sensitively and ignores namespaces, so `STRLEN($s) > 0` was missed
  while a namespaced user `strlen` (`Util\strlen($s)`, or an unqualified call
  where the namespace declares one) was reported and replaced by a string
  comparison that no longer calls it. custos matches case-insensitively and
  only calls resolving to the global functions (D1).
- **Cast decision uses native types only (custos diverges).** A subject
  known as `string` only through PHPDoc (`@return string` on an untyped
  method returning a `null` default, `@param string` on an untyped
  parameter) keeps the `(string)` cast: at run time it may be null, which
  `strlen()` counts as empty and `=== ''` does not (October CMS
  `Table::init()`, `!strlen($this->getConfig('dataSource'))`). Same policy
  as UnnecessaryCasting's PHPDoc-only types.

Recorded in `internal/testing/conformance/testdata/ea-divergences.json`.
