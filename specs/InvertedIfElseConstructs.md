---
id: InvertedIfElseConstructs
group: Control flow
kind: semantic
needs: [types]
php: { min: "", max: "" }
---

# InvertedIfElseConstructs

## Summary
An `if`/`else` whose condition is a negation (`!expr`, or `false === expr`)
reads backwards: swapping the two branches and dropping the negation makes the
main path positive and easier to follow.

## Detection
Visit every `else` branch `E` (the plain `else` keyword; an `else if (…)` is an
`else` whose body is a nested `if`).

- **D1** `E`'s body is a braced block `{ … }`. (`else if (…)`, `else stmt;` do
  not match.)
- **D2** The *partner* `P` of `E` is the last `elseif` branch of the owning
  `if` statement if it has any `elseif`, otherwise the `if` branch itself.
  For `if … else if … else`, the final `else` belongs to the nested `if`, so
  `P` is that nested `if`.
- **D3** `P`'s body is a braced block.
- **D4** `P`'s condition — the expression directly inside `if ( … )` /
  `elseif ( … )`, with **no** parenthesis stripping at this level — is one of:
  - **D4a** a logical-not `!X`. Let `X'` be `X` with all wrapping parentheses
    removed. `X'` must not be an `empty(...)` construct. `X'` can be anything
    else (call, variable, `isset(...)`, comparison, constant, …).
  - **D4b** a strict-identity comparison `L === R` where `L` or `R` is the
    constant `false` (case-insensitive, optional leading `\`); checked on
    `L` first. The other operand, parentheses stripped, is `X'`.
    Additionally, the type of `X'` must be known and consist of exactly one
    type (e.g. a call to a function declared `: bool`, a `true`/`false`
    constant). If the type cannot be resolved, or is a union of 2+ types
    (`?bool` = bool|null, untyped variables, mixed), there is no report.
- Conditions such as `(!$x)` (outer parentheses), `!$a && !$b`, `false == $x`,
  `$x !== false` do not match.

## Exceptions (no report)
- **E1** `if` without `else`.
- **E2** Either body unbraced.
- **E3** `!empty(...)` (D4a).
- **E4** `false === X` where `X` has an unknown or multi-type result (D4b).
- **E5** Parenthesised whole condition, compound conditions, non-strict
  comparisons.

## Report
- Range: the `else` keyword token of `E` (4 characters).
- Severity: info (weak warning).
- Message: `Negated condition with an else branch; swap the branches and drop the negation.`

## Fix
- **F1** New condition text `N`:
  - D4a: the verbatim source text of `X'` (parentheses that wrapped `X` are
    dropped: `!($a === $b)` → `$a === $b`);
  - D4b, when the single type of `X'` is `bool`: the verbatim source text of
    `X'` (`false === check()` → `check()`), since `false !== b` and `b` are
    the same test for a boolean;
  - D4b otherwise (a single non-`bool` type such as `int`, `true`, `false`):
    `Ltext !== Rtext` — the verbatim texts of the original left and right
    operands (not parenthesis-stripped) joined by ` !== ` (single spaces),
    keeping their order (the shorter form would change the meaning there).
- **F2** The whole content between the condition's parentheses of `P`
  (including any padding whitespace, e.g. `( ! ready() )`) is replaced so that
  it reads `(N)` — no space after `(` or before `)`.
- **F3** The two braced bodies (each including its `{` and `}`) are swapped:
  `P`'s body takes `E`'s former block and `E` takes `P`'s former block. Block
  contents are moved verbatim. Whitespace between `)` and `{` and between
  `else` and `{` may be normalised to a single space (upstream does so);
  either form is accepted because fix output is compared
  whitespace-collapsed.
- Other branches (`if` branch when `P` is an `elseif`, earlier `elseif`s) are
  untouched.

## Options
None.

## PHP versions
No gating. The D4b type check relies on declared return types (PHP 7.0+
syntax) or literal types; upstream fixtures use `: bool` / `: ?bool`
functions.

## Examples

```php
<?php
function ready(): bool { return true; }
function maybe(): ?bool { return null; }

if (!ready()) { wait(); }
<weak_warning descr="Negated condition with an else branch; swap the branches and drop the negation.">else</weak_warning> { go(); }

if ( ! ($left === $right) ) { differ(); }
<weak_warning descr="Negated condition with an else branch; swap the branches and drop the negation.">else</weak_warning> { same(); }

if ($mode === 1) { one(); }
elseif (!isset($cfg['x'])) { fallback(); }
<weak_warning descr="Negated condition with an else branch; swap the branches and drop the negation.">else</weak_warning> { useCfg(); }

if (FALSE === ready()) { wait(); }
<weak_warning descr="Negated condition with an else branch; swap the branches and drop the negation.">else</weak_warning> { go(); }

if (false === maybe()) { a(); } else { b(); }
if (!empty($list)) { a(); } else { b(); }
if ((!ready())) { a(); } else { b(); }
if (!ready()) { a(); } else b();
if (!ready()) a(); else { b(); }
if (!$p || !$q) { a(); } else { b(); }
if (!ready()) { a(); }
```

```php
<?php
function ready(): bool { return true; }
function maybe(): ?bool { return null; }

if (ready()) { go(); }
else { wait(); }

if ($left === $right) { same(); }
else { differ(); }

if ($mode === 1) { one(); }
elseif (isset($cfg['x'])) { useCfg(); }
else { fallback(); }

if (ready()) { go(); }
else { wait(); }

if (false === maybe()) { a(); } else { b(); }
if (!empty($list)) { a(); } else { b(); }
if ((!ready())) { a(); } else { b(); }
if (!ready()) { a(); } else b();
if (!ready()) a(); else { b(); }
if (!$p || !$q) { a(); } else { b(); }
if (!ready()) { a(); }
```

## Divergences
- **Positive D4b result (custos diverges):** upstream rewrites
  `false === x()` to `false !== x()`, which still reads as a negation — the
  very thing the rule complains about. When `x()` is known to be `bool`,
  custos emits the plain positive condition `x()` (F1); other single types
  keep the `!==` form, where dropping `false` would change the meaning.
- D4b's "exactly one resolved type" depends on the type engine: our engine may
  resolve more or fewer expressions than upstream. Recommendation: report only
  when the type is certain (declared non-nullable return type of a resolved
  function/method, a bool literal); otherwise skip.
- Alternative syntax (`if (!x): … else: … endif;`) is not covered by upstream
  fixtures. Recommendation: do not report (bodies are not braced blocks).
- **Empty else (custos diverges).** `if (!$ok) { … } else { }` (no
  statements, comments only) is not reported: swapping the branches would
  leave an empty `if` body — removing the empty `else` is the change to
  make, and the negation then reads naturally.
