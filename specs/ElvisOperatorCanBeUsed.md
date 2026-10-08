---
id: ElvisOperatorCanBeUsed
group: Language level migration
kind: syntax
needs: []
php: { min: "", max: "" }
---

# ElvisOperatorCanBeUsed

## Summary

A full ternary whose "then" branch repeats its condition (`$a ? $a : $b`) can
be written with the short ternary operator `$a ?: $b` (PHP 5.3+), which is
shorter and evaluates the condition once.

## Detection

- **D1** A ternary expression in its full form `C ? T : F` (a short ternary
  `C ?: F` is never reported).
- **D2** Let `C'` and `T'` be `C` and `T` with all surrounding parentheses
  stripped (any depth: `((($v)))` → `$v`). `F` must be present.
- **D3** `C'` and `T'` are equivalent:
  - both are the same node kind, and
  - for two simple variables: same name (`$v` vs `$v`);
  - otherwise structurally identical (same token sequence ignoring whitespace
    and comments), or identical source text.
- **D3a** `C` has no side effects: it contains no function, method or
  static call (including `|>`), `new`, `clone`, assignment, `++`/`--`,
  `include`/`require`, `eval`, `exit`/`die`, `print`, `throw`, `yield` or
  shell-exec backticks (closure and arrow-function bodies are not looked
  into). `next($it) ? next($it) : null` is not reported.
- **D4** Replacement text `R` = original source text of `C` (with its own
  parentheses, if any, preserved as written) + ` ?: ` + original source text of
  `F` (with its parentheses preserved). The text of `T` is dropped entirely.
- **Name case.** Wherever this rule compares two expressions for
  equivalence, the names PHP resolves case-insensitively — function and
  method names, class names in calls, `new`, `instanceof` and `::`
  accesses, and keywords — compare case-insensitively (`Cache::$map['k']`
  matches `cache::$map['k']`, `$o->Name()` matches `$o->name()`); variable,
  property and constant names stay case-sensitive.

## Exceptions (no report)

- **E1** Short ternaries.
- **E2** The repeated expression sits in the "else" branch (`$v ? 0 : $v`).
- **E4** The condition has side effects (D3a): `f() ? f() : $x`,
  `$s->fetch() ? $s->fetch() : []`, `$i++ ? $i++ : 0`.
- **E3** Condition and "then" branch differ in any way other than wrapping
  parentheses/whitespace, e.g. a negated condition (`!$v ? $v : 0`), or
  `isset($v) ? $v : 0` (the condition is `isset(...)`, not `$v`).

## Report

- Range: the whole ternary expression, from the first character of `C`
  (including `C`'s own opening parentheses) to the last character of `F`
  (including `F`'s parentheses). Parentheses enclosing the ternary itself are
  not part of the range.
- Severity: info (weak warning).
- Message: `Use the short ternary: '{R}'.`

## Fix

- **F1** Replace the reported ternary with `R` exactly: condition text, one
  space, `?:`, one space, else-branch text. No parentheses are added around the
  result.
  - `$p ? $p : 'n/a'` → `$p ?: 'n/a'`
  - `$p ? (($p)) : 'n/a'` → `$p ?: 'n/a'`
  - `(($p)) ? $p : 'n/a'` → `(($p)) ?: 'n/a'` (condition parentheses kept)
  - `$p->q ? $p->q : ($d)` → `$p->q ?: ($d)`

## Options

None.

## PHP versions

Upstream applies no gating (the short ternary exists since PHP 5.3, the
minimum custos supports).

## Examples

```php
<?php
$label  = <weak_warning descr="Use the short ternary: '$name ?: 'guest''.">$name ? $name : 'guest'</weak_warning>;
$limit  = <weak_warning descr="Use the short ternary: '$cfg['max'] ?: 10'.">$cfg['max'] ? ( $cfg['max'] ) : 10</weak_warning>;
$owner  = <weak_warning descr="Use the short ternary: '(($user)) ?: $fallback'.">(($user)) ? $user : $fallback</weak_warning>;
$title  = <weak_warning descr="Use the short ternary: '$page->title ?: (DEFAULT_TITLE)'.">$page->title ? $page->title : (DEFAULT_TITLE)</weak_warning>;

$keep1  = $name ?: 'guest';
$keep2  = $name ? 'guest' : $name;
$keep3  = !$name ? $name : 'guest';
$keep4  = isset($name) ? $name : 'guest';
$keep5  = getTitle() ? getTitle() : DEFAULT_TITLE;
```

```php
<?php
$label  = $name ?: 'guest';
$limit  = $cfg['max'] ?: 10;
$owner  = (($user)) ?: $fallback;
$title  = $page->title ?: (DEFAULT_TITLE);

$keep1  = $name ?: 'guest';
$keep2  = $name ? 'guest' : $name;
$keep3  = !$name ? $name : 'guest';
$keep4  = isset($name) ? $name : 'guest';
$keep5  = getTitle() ? getTitle() : DEFAULT_TITLE;
```

## Divergences

- custos diverges from upstream on side effects (D3a, E4). Upstream also
  reports conditions with calls such as `f() ? f() : $x`; the short ternary
  evaluates `f()` once where the original evaluated it twice, which changes
  behaviour for anything stateful (`next()`, `fetch()`, counters) and can
  return a different value. custos skips conditions that may have side
  effects. The EA conformance case still passes.
- Precedence: `R` is substituted without parentheses. If `C` is a
  low-precedence expression (assignment, `and`/`or`, `yield`, `print`) or the
  ternary is an operand of a tighter operator, the rewritten text can parse
  differently. Recommendation: wrap the result in parentheses only when the
  ternary itself was not standalone and its parent operator binds tighter than
  `?:`; otherwise emit `R` verbatim (matches upstream fixtures).
- Nested reportable ternaries (`$a ? $a : ($b ? $b : 0)`) produce two
  overlapping reports; recommendation: apply the outermost fix and re-run, or
  compose both (`$a ?: ($b ?: 0)`). Not covered by fixtures.
- **Name case (custos diverges).** Upstream compares the expressions
  textually, so operands that differ only in the case of a function, method
  or class name (`Stats::$n` vs `stats::$n`), which PHP treats as the same,
  are not recognised as equivalent. custos folds the case of those names
  (Detection, "Name case").
