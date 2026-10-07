---
id: UsingInclusionReturnValue
group: Code style
kind: syntax
needs: []
php: { min: "", max: "" }
---

# UsingInclusionReturnValue

## Summary
Using the value produced by `include`/`require` (a file that `return`s
something) couples code to file layout and hides dependencies. Prefer explicit
objects or functions; inclusion should be a standalone statement.

## Detection
- **D1** Visit every inclusion expression: `include`, `include_once`,
  `require`, `require_once` (with or without parentheses around the path).
- **D2** Going up from the inclusion through any number of `@` (silence)
  operators and parentheses, report when the first other ancestor is
  anything but an expression statement: `return require ...;`, an
  assignment right-hand side, a call argument, an operand of a comparison or
  other binary operator, an array element, a ternary branch, etc.
  (`$x = (require 'a.php');` and `$x = @include 'a.php';` are reported.)

## Exceptions (no report)
- **E1** `require 'x.php';` / `include_once('y.php');` as a statement of its own
  (top level or inside any block), also when silenced or parenthesized:
  `@include 'x.php';`, `(require 'x.php');`.

## Report
- Range: the inclusion expression itself, from the keyword to the end of its
  operand (closing `)` included when the path is parenthesized); not the
  surrounding statement, not the `;`.
- Severity: info.
- Message: `Avoid relying on the value returned by an included file.`

## Fix
None.

## Options
None.

## PHP versions
None.

## Examples

```php
<?php
require_once __DIR__ . '/bootstrap.php';
include('helpers.php');

function settings(array $env) {
    $base = <weak_warning descr="Avoid relying on the value returned by an included file.">require __DIR__ . '/base.php'</weak_warning>;
    merge(<weak_warning descr="Avoid relying on the value returned by an included file.">include_once 'extra.php'</weak_warning>);
    $env['db'] = <weak_warning descr="Avoid relying on the value returned by an included file.">require_once('db.php')</weak_warning>;
    if (true !== <weak_warning descr="Avoid relying on the value returned by an included file.">include 'optional.php'</weak_warning>) {
        log_missing();
    }
    return <weak_warning descr="Avoid relying on the value returned by an included file.">include('defaults.php')</weak_warning>;
}
```

## Divergences
- **D2 — custos diverges from upstream.** Upstream checks only the direct
  parent, so `@include 'optional.php';` (and a parenthesized inclusion used
  as a statement) is reported although its value is discarded. custos looks
  through `@` and parentheses before deciding whether the inclusion is a
  statement of its own.
