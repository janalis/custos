---
id: SuspiciousFunctionCalls
group: Probable bugs
kind: syntax
needs: []
php: { min: "", max: "" }
---

# SuspiciousFunctionCalls

## Summary
A string comparison function called with the same expression as both
operands always yields "equal" — typically a copy-paste slip where one side
should have been a different variable.

## Detection
- **D1** A function call (not a method call) whose name part is exactly one of
  `strcmp`, `strncmp`, `strcasecmp`, `strncasecmp`, `strnatcmp`,
  `strnatcasecmp`, `substr_compare`, `hash_equals`, compared
  case-insensitively against the name as written (`STRCMP`, `Hash_Equals`
  match, as PHP function names are case-insensitive), and the call resolves
  to that global function: `\strcmp(...)` and an unqualified call in the
  global namespace match; an unqualified call inside a namespace matches
  only when no function of that name is declared in the namespace or
  imported with `use function`; a qualified `Foo\strcmp(...)` does not
  match (custos diverges).
- **D2** The call has at least two arguments.
- **D3** The first and second arguments are equivalent: same node kind and,
  for simple variables, the same name; otherwise the same token sequence
  ignoring whitespace and comments, or identical source text. Parentheses are
  significant (`$a` vs `($a)` are not equivalent). No purity check is made, so
  `strcmp(next($it), next($it))` is reported.
- **Name case.** Wherever this rule compares two expressions for
  equivalence, the names PHP resolves case-insensitively — function and
  method names, class names in calls, `new`, `instanceof` and `::`
  accesses, and keywords — compare case-insensitively (`Cache::$map['k']`
  matches `cache::$map['k']`, `$o->Name()` matches `$o->name()`); variable,
  property and constant names stay case-sensitive.

## Exceptions (no report)
- **E1** Fewer than two arguments.
- **E2** Different first and second arguments (later arguments are ignored).
- **E3** Other functions (`strcoll`, `levenshtein`, …) and method calls.

## Report
- Range: the whole call, from the start of the name (including any namespace
  qualifier) to the closing `)`.
- Severity: error.
- Message: `Both compared strings are the same expression; one of them is probably wrong.`

## Fix
None.

## Options
None.

## PHP versions
No gating.

## Examples

```php
<?php
function verify(string $token, string $expected, array $row) {
    $a = <error descr="Both compared strings are the same expression; one of them is probably wrong.">hash_equals($token, $token)</error>;
    $b = <error descr="Both compared strings are the same expression; one of them is probably wrong.">\strcasecmp($row['name'], $row[ 'name' ])</error>;
    $c = <error descr="Both compared strings are the same expression; one of them is probably wrong.">substr_compare($token, $token, 0, 4)</error>;
    $d = strncmp($token, $expected, 4);
    $e = strnatcmp($token, ($token));
    return [$a, $b, $c, $d, $e];
}
```

## Divergences
- **Name case (custos diverges).** Upstream matches the function name
  case-sensitively, so `STRCMP($a, $a)` or `Hash_Equals($t, $t)` escape the
  check even though PHP calls the same function. custos compares the name
  case-insensitively (D1).
- **Callee resolved (custos diverges).** Upstream accepts any namespace
  qualifier and does not resolve the call, so a user function
  `App\strcmp($a, $a)` (declared in the namespace, imported, or written
  qualified) is reported although its arguments may legitimately repeat.
  custos reports only calls that reach the builtin (D1).
- **Name case (custos diverges).** Upstream compares the expressions
  textually, so operands that differ only in the case of a function, method
  or class name (`Stats::$n` vs `stats::$n`), which PHP treats as the same,
  are not recognised as equivalent. custos folds the case of those names
  (Detection, "Name case").
