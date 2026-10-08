---
id: GetClassUsage
group: Language level migration
kind: semantic
needs: [types]
php: { min: "7.1", max: "" }
---

# GetClassUsage

## Summary
Since PHP 7.2, passing `null` to `get_class()` is an error-level misuse (it no
longer silently falls back to the calling class). This rule flags
`get_class(x)` calls whose single argument may be `null` and that are not
preceded by any null check of that same expression in the enclosing function.

## Detection
- **D1** Node: a function call (not a method/static call) whose name part is
  `get_class` (case-insensitive: `Get_Class` matches) and that resolves to the global
  `get_class()` with PHP's runtime rules: `get_class(...)` and
  `\get_class(...)` match; `Some\get_class(...)`, a call imported through
  `use function Some\get_class;`, or an unqualified call in namespace `Some`
  where `Some\get_class()` is a known function do not.
- **D2** The call has exactly one argument `A`. Zero or two-plus arguments are
  never reported.
- **D3** The configured PHP language level is **7.1 or higher**. Below 7.1
  the rule is silent.
- **D4** `A` may be null, i.e. at least one of:
  - **D4a** the inferred type of `A` (ignoring unknown parts) contains
    `null`. Typical sources: the literal `null` (any case, optionally
    `\`-qualified), a variable/parameter with a nullable declaration
    (`?Foo $p`, `Foo $p = null`, `Foo|null $p`), a parameter whose default is
    `null`, a call to a function/method declared to return `?T` / `T|null` /
    `null`, a nullable typed property, a local variable assigned `null`;
  - **D4b** `A` is a plain variable `$n`, and the nearest enclosing function
    (named function, method or closure) declares a parameter named `n` whose
    default value is the constant `null` (any case, optionally `\`-qualified).
    This holds regardless of the parameter's type declaration and regardless
    of any reassignment of `$n` before the call.
- **D5** `A` is not "already null-checked" (see the check below). If D4
  holds and the check fails, report.

### Null-check search (used by D5)
Performed only when the call sits inside a function/method/closure body
(the **nearest** enclosing one). At file/top level there is no search and the
call is always reported when D4 holds.

1. Collect every node inside that function's body (at any depth, including
   inside nested closures) that is of the **same node kind** as `A` and is
   equivalent to `A`: for variables, the same variable name; for other
   expressions, structurally identical or identical source text.
2. Keep only those that come **before** `A` in source order (pre-order tree
   traversal). Order is purely textual — branches, loops and reassignments
   are not considered.
3. `A` counts as checked if any kept occurrence `C` satisfies one of:
   - **C1** `C` is a direct argument of `isset(...)` or `empty(...)`;
   - **C2** `C` is a direct operand of an `instanceof` expression (either
     side);
   - **C3** `C` is a direct operand of `==`, `!=`, `<>`, `===` or `!==`
     whose other operand is the constant `null` (any case);
   - **C4** `C` is used as a boolean operand: after climbing any number of
     enclosing parentheses, its parent is the condition of an `if`,
     `elseif`, `while` or `do … while`; or a logical-not `!`; or either
     operand of `&&`, `||`, `and`, `or`; or the condition (not a branch) of a
     full ternary `c ? x : y` (the short form `?:` does not count).
- **Name case.** Wherever this rule compares two expressions for
  equivalence, the names PHP resolves case-insensitively — function and
  method names, class names in calls, `new`, `instanceof` and `::`
  accesses, and keywords — compare case-insensitively (`Cache::$map['k']`
  matches `cache::$map['k']`, `$o->Name()` matches `$o->name()`); variable,
  property and constant names stay case-sensitive.

## Exceptions (no report)
- **E1** PHP level below 7.1.
- **E2** Argument type known and not containing `null`, and not a
  null-defaulted parameter (e.g. `$s = ''`, `string $s`, `$this`, `new X`).
- **E3** Argument type entirely unknown (nothing remains after dropping
  unknown parts) and not a null-defaulted parameter.
- **E4** An equivalent occurrence earlier in the same function body passes
  C1–C4, even when that check is in an unrelated branch or was followed by a
  reassignment.
- **E5** Checks that are not recognised: `is_null($n)`, `$n ?? …`,
  `isset($n->prop)` (the occurrence's parent is the member access, not
  `isset`), comparisons with anything other than the `null` constant
  (`$n === false`), occurrences after the call.

## Report
- Range: the whole call expression, from the start of the function name
  (including any leading `\` or namespace qualifier) to the closing `)`.
- Severity: warning.
- Message: `get_class() rejects null on PHP 7.2+; guard the argument.`

## Fix
None.

## Options
None.

## PHP versions
- Active only when the language level is >= 7.1 (the upstream threshold;
  the PHP behaviour change itself is in 7.2).
- Upstream fixture is run at level 7.1. Tests with no explicit level run below
  7.1 in upstream's harness and would produce no reports.

## Examples
Language level 7.4:

```php
<?php

function describe($entity = null, $label = 'x', ?Order $order = null, Invoice $invoice = NULL) {
    <warning descr="get_class() rejects null on PHP 7.2+; guard the argument.">\get_class(NULL)</warning>;
    <warning descr="get_class() rejects null on PHP 7.2+; guard the argument.">get_class($entity)</warning>;
    <warning descr="get_class() rejects null on PHP 7.2+; guard the argument.">get_class($order)</warning>;

    get_class($label);
    get_class($entity, 1);

    if (empty($invoice)) {
        return;
    }
    get_class($invoice);

    while ($order instanceof Order) {
        get_class($order);
    }
}

function report(?Order $first, ?Order $second) {
    $tag = $first ? 'a' : 'b';
    echo get_class($first);
    echo <warning descr="get_class() rejects null on PHP 7.2+; guard the argument.">get_class($second)</warning>;
    if (null != $second) {}
}

$loose = null;
<warning descr="get_class() rejects null on PHP 7.2+; guard the argument.">get_class($loose)</warning>;
```

Language level 7.0 — nothing is reported:

```php
<?php
function legacy($item = null) {
    return get_class($item);
}
```

## Divergences
- Arrow functions (`fn($m = null) => get_class($m)`): upstream treats the
  arrow function as the enclosing scope but has no block body to search;
  this path throws internally and nothing gets reported. Recommendation:
  treat an arrow function as having no preceding checks except those inside
  its own expression body that come before `A`, and report otherwise. No
  fixture covers it.
- The null-check search is order-based and branch-insensitive (a check in an
  unrelated branch suppresses later calls). Kept on purpose: requiring a
  dominating check would also have to understand early exits
  (`if ($x === null) { return; }`), `assert()`, `?? throw` and narrowing
  through helper calls; without full flow narrowing it would mostly add
  false positives for a small gain.
- Namespaced `get_class` (custos diverges from upstream). Upstream matches
  any function call named `get_class`, so a user-defined
  `Some\get_class($x)` (whose null handling is its own business) is
  reported. custos only considers calls that resolve to the global
  `get_class()` (D1).
- **Function-name case (custos diverges):** upstream matches `get_class`
  case-sensitively, missing `Get_Class($maybeNull)`. custos matches any case
  (D1).
- **Name case (custos diverges).** Upstream compares the expressions
  textually, so operands that differ only in the case of a function, method
  or class name (`Stats::$n` vs `stats::$n`), which PHP treats as the same,
  are not recognised as equivalent. custos folds the case of those names
  (Detection, "Name case").
- **Flow-aware D4b (custos diverges).** A parameter with a `null` default
  counts as possibly null only when the inferred type at the call is
  unknown; when inference proves it non-null (`$t = $t ?: $this;`) the call
  is not reported. A type guard `is_object($x)`, `is_a($x, …)` or
  `is_subclass_of($x, …)` earlier in the function counts as a null check
  (C5). Found on Magento (`setObject($object = null) { if
  (is_object($object)) { get_class($object) … }`).
