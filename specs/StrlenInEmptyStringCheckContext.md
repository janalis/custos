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
(`strlen($s) > 0`, `!mb_strlen($s)`, `if (strlen($s))`) is indirect and
slower than comparing with the empty string. Suggest an identity comparison
with `''` instead, casting to string when the value is not known to be a
string.

## Detection
- **D1** A plain function call that resolves to the global function
  `strlen` or `mb_strlen` — name compared case-insensitively as PHP does
  (`StrLen`, `\MB_STRLEN` match); written with a single leading `\`, or
  unqualified with no function of that name declared or imported in the
  current namespace (`Ns\strlen(...)` and shadowing user functions are not
  matched) — with at least one argument
  (extra arguments such as an encoding are allowed and ignored). Call it
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
  - The operand order follows the global comparison style setting
    (`comparisonStyle`, default `regular`):
    - regular: `{v} {op} ''`
    - yoda: `'' {op} {v}`
  Single spaces around `{op}`; `''` is two single quotes. The upstream
  fixture runs in **yoda** style, so conformance expects `'' === $s`,
  `'' !== (string)$n`.
- Examples (regular): `strlen($name) > 0` → `$name !== ''`;
  `!mb_strlen($id)` with unknown `$id` → `(string)$id === ''`.
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
Yoda style:

```php
<?php
function probe(string $name, ?string $alias, int $count, $raw, array $bag)
{
    if (<weak_warning descr="Compare with an empty string instead: ''' !== $name'.">strlen($name)</weak_warning>) {}
    while (<weak_warning descr="Compare with an empty string instead: ''' === $name'.">!mb_strlen($name, 'UTF-8')</weak_warning>) {}
    if (!(<weak_warning descr="Compare with an empty string instead: ''' !== (string)$alias'.">strlen($alias)</weak_warning>)) {}
    $a = <weak_warning descr="Compare with an empty string instead: ''' !== (string)$raw'.">mb_strlen($raw)</weak_warning> ? 'y' : 'n';
    $b = $count > 2 and <weak_warning descr="Compare with an empty string instead: ''' !== trim($raw)'.">strlen(trim($raw))</weak_warning>;

    $c = <weak_warning descr="Compare with an empty string instead: ''' === $name'.">0 === mb_strlen($name)</weak_warning>;
    $d = <weak_warning descr="Compare with an empty string instead: ''' !== (string)$count'.">strlen($count) <> 0</weak_warning>;
    $e = <weak_warning descr="Compare with an empty string instead: ''' === (string)$alias'.">strlen($alias)  ==  0</weak_warning>;
    $f = <weak_warning descr="Compare with an empty string instead: ''' !== $name'.">mb_strlen($name) > 0</weak_warning>;
    $g = <weak_warning descr="Compare with an empty string instead: ''' !== (string)$bag['k']'.">strlen($bag['k']) >= 1</weak_warning>;
    $h = <weak_warning descr="Compare with an empty string instead: ''' === $name'.">strlen($name) < 1</weak_warning>;

    $i = 0 < strlen($name);
    $j = strlen($name) > 1;
    $k = (strlen($name)) > 0;
    $l = strlen($name) ?: 5;
    return strlen($name);
}

if (<weak_warning descr="Compare with an empty string instead: ''' !== (string)$argv[1]'.">strlen($argv[1])</weak_warning>) {}
```

```php
<?php
function probe(string $name, ?string $alias, int $count, $raw, array $bag)
{
    if ('' !== $name) {}
    while ('' === $name) {}
    if (!('' !== (string)$alias)) {}
    $a = '' !== (string)$raw ? 'y' : 'n';
    $b = $count > 2 and '' !== trim($raw);

    $c = '' === $name;
    $d = '' !== (string)$count;
    $e = '' === (string)$alias;
    $f = '' !== $name;
    $g = '' !== (string)$bag['k'];
    $h = '' === $name;

    $i = 0 < strlen($name);
    $j = strlen($name) > 1;
    $k = (strlen($name)) > 0;
    $l = strlen($name) ?: 5;
    return strlen($name);
}

if ('' !== (string)$argv[1]) {}
```

Regular style (default), same input lines `$c`, `$d`, `$a` become
`$name === ''`, `(string)$count !== ''`, `(string)$raw !== '' ? 'y' : 'n'`.

## Divergences
- **Parenthesised argument (custos diverges).** Upstream inserts `A`'s
  text unparenthesised, so low-precedence arguments produce wrong code:
  `strlen($p ?: $q) > 0` → `'' !== $p ?: $q`; with a cast,
  `(string)$p . $q` / `(string)$p + $q` binds the cast to `$p` only.
  custos wraps `A` in parentheses when needed (F1a): `($p ?: $q) !== ''`,
  `(string)($raw + 1) !== ''`. Not covered by upstream fixtures.
- Replacing a bare `L` inside a `&&`/`||`/`and`/`or` operand keeps
  precedence correct for `===`/`!==` in all those contexts, so no change is
  needed there.
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
