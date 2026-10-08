---
id: GetDebugTypeCanBeUsed
group: Language level migration
kind: semantic
needs: [names, stubs]
php: { min: "8.0", max: "" }
---

# GetDebugTypeCanBeUsed

## Summary

The hand-rolled idiom "class name for objects, `gettype()` otherwise",
written as a ternary over `is_object()`, is what PHP 8.0's
`get_debug_type()` provides in one call. Suggest the built-in.

## Detection

- **D1** Node: a full ternary `C ? T : F` (the short form `C ?: F` is never
  considered).
- **D2** `C` is directly a function call (not wrapped in parentheses, not
  negated) whose name part is `is_object` (case-insensitive, as PHP compares
  function names) with
  exactly one argument `X`.
- **D3** The language level is **8.0 or higher**.
- **D4** `T` is directly a function call named `get_class` (any case) with
  exactly one argument, and that argument is equivalent to `X`.
- **D5** `F` is directly a function call named `gettype` (any case) with exactly
  one argument, and that argument is equivalent to `X`.
- **D6** Each of the three calls (`is_object`, `get_class`, `gettype`) must
  resolve to the **global** built-in function of that name: unqualified in
  the global namespace, `\`-qualified, or unqualified inside a namespace
  where it falls back to the global function. If a call resolves to some
  other function (e.g. `use function Lib\gettype;`, or a function
  `is_object` declared in the current namespace) or cannot be resolved at
  all, no report.

"Equivalent" means the same node kind and structurally identical (ignoring
whitespace and comments) or identical source text; for plain variables, the
same name.

- **Name case.** Wherever this rule compares two expressions for
  equivalence, the names PHP resolves case-insensitively — function and
  method names, class names in calls, `new`, `instanceof` and `::`
  accesses, and keywords — compare case-insensitively (`Cache::$map['k']`
  matches `cache::$map['k']`, `$o->Name()` matches `$o->name()`); variable,
  property and constant names stay case-sensitive.

## Exceptions (no report)

- **E1** Short ternary `?:`.
- **E2** Any branch not of the exact shape above: a parenthesised branch or
  condition, `!is_object(...)`, swapped branches (`gettype` in the true
  branch), a literal string in the false branch, different arguments
  between the calls, extra arguments.
- **E3** Language level below 8.0.
- **E4** Calls bound to non-global functions (D6).

## Report

- Range: the whole ternary expression, from the start of `C` to the end of
  `F` (enclosing parentheses, if any, are not included).
- Severity: info (fixtures tag it `weak_warning`).
- Message: `Use '{replacement}' instead (scalar type names differ from
  gettype()).` where `{replacement}` is `{prefix}get_debug_type({X})`:
  - `{prefix}` is the namespace qualifier written in front of the
    `is_object` call: empty when unqualified, `\` when written
    `\is_object(...)`;
  - `{X}` is the verbatim source text of the `is_object` argument.

## Fix

- **F1** No automatic fix. `get_debug_type()` does not return the same
  strings as the idiom: `int`/`integer`, `float`/`double`, `bool`/`boolean`,
  `null`/`NULL`, `resource (stream)`/`resource`, and a shortened name for
  anonymous classes. Code that compares, stores or displays the result can
  change behaviour, so the rewrite is left to the developer.

## Options

None.

## PHP versions

Active only at language level >= 8.0. The upstream fixture runs at 8.0;
tests without an explicit level would run below 7.1 and see no report.

## Examples

Language level 8.1:

```php
<?php
namespace Shop;

function label(object|string $payload, array $rows): array {
    return [
        <weak_warning descr="Use 'get_debug_type($rows[0])' instead (scalar type names differ from gettype()).">is_object($rows[0]) ? get_class($rows[0]) : gettype($rows[0])</weak_warning>,
        'kind: ' . (<weak_warning descr="Use '\get_debug_type($payload)' instead (scalar type names differ from gettype()).">\is_object($payload) ? \get_class($payload) : gettype($payload)</weak_warning>),
        is_object($payload) ? get_class($payload) : gettype($rows),
        is_object($payload) ? gettype($payload) : get_class($payload),
        is_object($payload) ? get_class($payload) : 'scalar',
        is_object($payload) ?: gettype($payload),
        (is_object($payload)) ? get_class($payload) : gettype($payload),
    ];
}
```

Language level 7.4 — nothing is reported:

```php
<?php
$t = is_object($v) ? get_class($v) : gettype($v);
```

## Divergences

- custos diverges from upstream on the fix (F1). Upstream replaces the
  ternary with `get_debug_type()`, but the two produce different strings
  for integers, floats, booleans, null, resources and anonymous classes
  (e.g. `integer` becomes `int`), so a later comparison such as
  `$t === 'integer'` or a stored/logged value changes. custos keeps the
  report as advice and offers no fix. The EA case whose expected output
  applies the rewrite is listed in `testdata/ea-divergences.json`.
- **Function-name case (custos diverges):** upstream matches `is_object`,
  `get_class` and `gettype` case-sensitively, missing
  `Is_Object($x) ? GET_CLASS($x) : GetType($x)`. custos matches any case
  (D2, D4, D5).
- **Name case (custos diverges).** Upstream compares the expressions
  textually, so operands that differ only in the case of a function, method
  or class name (`Stats::$n` vs `stats::$n`), which PHP treats as the same,
  are not recognised as equivalent. custos folds the case of those names
  (Detection, "Name case").
