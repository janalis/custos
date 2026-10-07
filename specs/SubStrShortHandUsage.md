---
id: SubStrShortHandUsage
group: Performance
kind: syntax
needs: []
php: { min: "", max: "" }
---

# SubStrShortHandUsage

## Summary
`substr()` / `mb_substr()` accept a negative length ("stop N characters before
the end") and treat a missing length as "up to the end". Computing the length
as `strlen($s) - something` is therefore either replaceable by a negative
constant, or entirely unnecessary.

## Detection
- **D1** A plain function call that resolves to the global function (names
  compared case-insensitively, as PHP does; `\` and a global `use function`
  import are fine, but a same-named function declared in the current
  namespace, one imported from another namespace, or a qualified non-global
  name such as `Ns\substr` does not count): `substr` or `mb_substr`. Call it
  `S`, with arguments `subject` (1st), `start` (2nd), `length` (3rd) and
  optional `encoding` (4th).
- **D2** `S` has **3** arguments, or **4** arguments when `S` is
  `mb_substr` (plain `substr` has no 4th parameter, see E1b).
- **D3** `length` is directly (no wrapping parentheses) a binary subtraction
  `L - R` (operator `-`).
- **D4** `L` is a plain function call resolving (as in D1) to the global
  function that counts in the same unit as `S`: `strlen` when `S` is
  `substr` (bytes), `mb_strlen` when `S` is `mb_substr` (characters). It has
  **exactly 1** argument, and that argument is equivalent to `subject` (see
  "Equivalence").
- Then:
  - **D5** *Drop*: `R` is equivalent to `start` → report "length is
    unnecessary" on `length`.
  - **D6** Otherwise, when **both** `start` and `R` are integer number
    literals — a decimal literal, optionally preceded by a unary minus — whose
    source text parses as a (32-bit) decimal integer:
    let `d = int(start) - int(R)`.
    - `d < 0` → *Simplify*: report "length can be `d`" on `length`.
    - `d >= 0` → *Drop*: report "length is unnecessary" on `length`.
    Texts that do not parse as a plain decimal integer (hex `0x2`, octal-like
    leading zero is fine as decimal, binary `0b1`, floats `1.0`, numbers with
    `_` separators, `- 2` with a space after the minus, values out of 32-bit
    range) → no report.
  - Any other `start`/`R` combination (variables, calls, constants, …, not
    equivalent to each other) → no report.

### Equivalence
Two expressions are equivalent when they are of the same node kind and are
structurally identical (same tokens, ignoring whitespace and comments), or
have exactly the same source text. For two simple variables, compare names
only (`$s` vs `$s`).
- **Name case.** Wherever this rule compares two expressions for
  equivalence, the names PHP resolves case-insensitively — function and
  method names, class names in calls, `new`, `instanceof` and `::`
  accesses, and keywords — compare case-insensitively (`Cache::$map['k']`
  matches `cache::$map['k']`, `$o->Name()` matches `$o->name()`); variable,
  property and constant names stay case-sensitive.

## Exceptions (no report)
- **E1** 2 or ≥ 5 arguments.
- **E1b** `substr` called with 4 arguments (`substr($s, 0, strlen($s) - 1,
  'x')`): the call itself is broken (too many arguments), so no rewrite is
  offered.
- **E2** `length` wrapped in parentheses, or using another operator (`+`, `*`).
- **E3** `L` measures something other than `subject` (`strlen($other) - 1`),
  or takes 2 arguments (`mb_strlen($s, 'UTF-8') - 1`).
- **E3b** `L` counts in the other unit (`mb_substr($s, 0, strlen($s) - 2)`,
  `substr($s, $k, mb_strlen($s) - $k)`): on multibyte strings the byte and
  character counts differ, so neither rewrite would be equivalent.
- **E4** `L - R` reversed (`2 - strlen($s)`).
- **E5** Non-literal, non-equivalent `start`/`R` (`substr($s, $i, strlen($s) - 3)`,
  `substr($s, 0, strlen($s) - strlen($p))`).

## Report
- Range: the `length` argument (the whole `L - R` expression).
- Severity: warning for both cases (Drop is shown upstream with the "unused
  symbol" look, severity still warning).
- Messages:
  - Simplify: `Pass '{d}' as the length instead.` (`{d}` is the decimal
    value, e.g. `-3`).
  - Drop: `The length '{length}' is unnecessary; remove it.` (`{length}` is
    the verbatim text of the 3rd argument).

## Fix
- **F1** *Simplify*: replace `length` with the decimal text of `d`
  (e.g. `-3`).
- **F2** *Drop*, 3 arguments: rewrite the argument list (the content between
  the call's parentheses) as `{subject}, {start}` — verbatim texts joined by
  `, `. Whitespace/newlines between the `(` and the first argument and
  between the last argument and `)` are kept; everything in between
  (original separators, comments, the removed argument) is replaced.
- **F3** *Drop*, 4 arguments (`mb_substr` only): rewrite the argument list as
  `{subject}, {start}, null, {encoding}` (the length becomes `null`, which
  means "to the end").

Examples:
- `substr($s, 0, strlen($s) - 3)` → `substr($s, 0, -3)`
- `substr($s, 2, strlen($s) - 2)` → `substr($s, 2)`
- `mb_substr($s, $n, mb_strlen($s) - $n, 'UTF-8')` → `mb_substr($s, $n, null, 'UTF-8')`

## Options
None.

## PHP versions
No gating (passing `null` as length to `mb_substr` is valid on every version
that has it; for plain `substr` the 4-argument form does not exist, but the
rule does not care). Upstream fixture runs at the test default level.

## Examples

```php
<?php
function cut($name, $head, $k) {
    $a = substr($name, 0, <warning descr="Pass '-4' as the length instead.">strlen($name) - 4</warning>);
    $b = mb_substr($name, 3, <warning descr="Pass '-2' as the length instead.">mb_strlen($name) - 5</warning>, 'UTF-8');
    $c = \substr($name, 7, <warning descr="The length 'strlen($name) - 2' is unnecessary; remove it.">strlen($name) - 2</warning>);
    $d = mb_substr($name, $k, <warning descr="The length 'mb_strlen($name) - $k' is unnecessary; remove it.">mb_strlen($name) - $k</warning>);
    $e = mb_substr(
        $name,
        strlen($head),
        <warning descr="The length 'mb_strlen($name) - strlen($head)' is unnecessary; remove it.">mb_strlen($name) - strlen($head)</warning>,
        'UTF-8'
    );

    $f = substr($name, $k, strlen($name) - 1);
    $g = substr($name, 0, strlen($head) - 1);
    $h = substr($name, 0, (strlen($name) - 1));
    $i = substr($name, 0x1, strlen($name) - 2);
    $j = substr($name, 0, mb_strlen($name, 'UTF-8') - 1);
    $m = mb_substr($name, 0, strlen($name) - 2);
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i, $j, $m];
}
```

```php
<?php
function cut($name, $head, $k) {
    $a = substr($name, 0, -4);
    $b = mb_substr($name, 3, -2, 'UTF-8');
    $c = \substr($name, 7);
    $d = mb_substr($name, $k);
    $e = mb_substr(
        $name, strlen($head), null, 'UTF-8'
    );

    $f = substr($name, $k, strlen($name) - 1);
    $g = substr($name, 0, strlen($head) - 1);
    $h = substr($name, 0, (strlen($name) - 1));
    $i = substr($name, 0x1, strlen($name) - 2);
    $j = substr($name, 0, mb_strlen($name, 'UTF-8') - 1);
    $m = mb_substr($name, 0, strlen($name) - 2);
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i, $j, $m];
}
```

## Divergences
- **Byte vs character length mixing (custos diverges):** upstream accepts
  `strlen` and `mb_strlen` interchangeably for both `substr` and `mb_substr`,
  so `mb_substr($s, 0, strlen($s) - 2)` becomes `-2` and `substr($s, $k,
  mb_strlen($s) - $k)` loses its length. On multibyte input the byte count
  exceeds the character count, so both rewrites return a different string.
  custos requires the length function to match the variant (D4, E3b) and stays
  silent on mixed pairs.
- **Negative literal start:** `substr($s, -3, strlen($s) - 1)` gives
  `d = -2` → simplify to `-2`, which is not equivalent (original returns the
  last 3 characters). Recommendation: only apply D6 when `start` is
  non-negative; no upstream fixture covers negative starts.
- **4-argument `substr` (custos diverges):** plain `substr` has no 4th
  parameter, yet upstream still reports such calls and its fix keeps the
  stray argument after a new `null`. The call is already an argument-count
  error, so custos stays silent on 4-argument `substr` (E1b) and only
  handles the 4-argument form for `mb_substr`.
- **Function-name matching — custos diverges from upstream.** Upstream matches
  `substr`/`mb_substr` and `strlen`/`mb_strlen` by the name as written
  (case-sensitive, any namespace qualifier, no resolution), so a differently
  cased call such as `SubStr($s, 0, strlen($s) - 2)` is missed while a
  namespaced or imported user function of the same name is reported (and
  rewritten) as if it were the builtin. custos matches case-insensitively and
  only calls that reach the global function.
- **Name case (custos diverges).** Upstream compares the expressions
  textually, so operands that differ only in the case of a function, method
  or class name (`Stats::$n` vs `stats::$n`), which PHP treats as the same,
  are not recognised as equivalent. custos folds the case of those names
  (Detection, "Name case").
- **Simplify only when `d >= -2` (custos diverges).** For a string shorter
  than `R` the original length `strlen($s) - R` is negative and PHP still
  keeps characters up to `2·strlen($s) - R`, while the suggested constant
  `d` keeps none: `substr('abcde', 3, strlen('abcde') - 6)` is `'d'`,
  `substr('abcde', 3, -3)` is `''`. The two differ for some length in
  `((R + start) / 2, R)`, which contains an integer exactly when
  `d <= -3`. custos reports *Simplify* for `d` = -1 or -2 only and stays
  silent below.
