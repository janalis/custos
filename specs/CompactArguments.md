---
id: CompactArguments
group: Probable bugs
kind: syntax
needs: []
php: { min: "", max: "" }
---

# CompactArguments

## Summary

`compact('name')` silently skips (or, since PHP 7.3, warns about) variables
that do not exist at that point. A name that never appeared earlier in the
function is usually a typo or a variable assigned too late.

## Detection

Visit every function call.

- **D1** The call resolves to the global function `compact` (name compared
  case-insensitively, as PHP does: `Compact()` qualifies): `\compact()` and
  an unqualified call in the global namespace match; an unqualified call in
  a namespace matches only when no `compact` function is declared in that
  namespace or imported with `use function`; a qualified `Foo\compact()`
  does not (custos diverges).
- **D2** At least one argument.
- **D3** The call is inside a function, method or closure: `S` is the
  nearest enclosing function, method or closure. Arrow functions are not a
  scope of their own (they capture the parent scope by value), so they are
  looked through: a call inside `fn() => compact('x')` uses the function,
  method or closure around the arrow function as `S`, and the arrow
  function's own parameters are earlier variable occurrences of `S` (D5).
  Top-level calls (including inside an arrow function at top level) are
  ignored.
- **D4** Collect *candidate names* from the arguments: each argument that is
  itself a string literal (single- or double-quoted) **without
  interpolation** and with non-empty raw content contributes its raw content
  as a name (the content is taken verbatim: `'$total'` gives the name
  `$total`, with the dollar sign). Other arguments (variables, arrays,
  interpolated strings, concatenations) are ignored. When the same name
  occurs several times, only its **last** occurrence is kept as the
  candidate's position.
- **D5** Build the set of *known* names:
  - the names (without `$`) of all parameters of `S`;
  - the names of every variable occurrence (any `$name` token: reads,
    writes, `$this`, `global`/`static` declarations, closure `use (...)`
    variables, variables inside nested closures/functions) located inside
    `S` and appearing in source order **before** this `compact` call
    expression starts. Occurrences inside the `compact(...)` call itself or
    after it do not count.
- **D6** Report each candidate name that is not in the known set.

## Exceptions (no report)

- **E1** `compact()` at file top level.
- **E2** Names of parameters, or of variables mentioned anywhere earlier in the
  scope (no flow analysis: a variable assigned in a branch that is not taken,
  or only read before, still counts as known).
- **E3** Interpolated strings and non-literal arguments.
- **E4** Earlier duplicates of a reported name (only the last occurrence is
  highlighted).

## Report

- Range: the string literal argument (including its quotes) of the last
  occurrence of the name.
- Severity: error.
- Message: `Variable '${name}' may be undefined when compact() runs.` where
  `{name}` is the raw literal content (so `'$total'` yields `'$$total'`).

## Fix

None.

## Options

None.

## PHP versions

None.

## Examples

```php
<?php
function summary($net, $tax)
{
    $gross = $net + $tax;
    return compact(
        'net',
        "tax",
        'gross',
        'discount',
        <error descr="Variable '$discount' may be undefined when compact() runs.">'discount'</error>,
        <error descr="Variable '$$gross' may be undefined when compact() runs.">'$gross'</error>,
        "{$gross}",
        $net
    );
}

function lateAssignment($unit)
{
    $payload = compact('unit', <error descr="Variable '$count' may be undefined when compact() runs.">'count'</error>);
    $count = 3;
    return $payload;
}

class Report
{
    public function build()
    {
        $title = 'Q3';
        $render = function () use ($title) {
            return compact('title', <error descr="Variable '$footer' may be undefined when compact() runs.">'footer'</error>);
        };
        return compact('title', 'render');
    }
}

$top = compact('nothing');
```

## Divergences

- **Arrow functions (custos diverges):** upstream treats an arrow function
  as a separate scope, so `fn() => compact('total')` reports `total` even
  when the enclosing function defines `$total`, which the arrow function
  captures automatically. custos looks through arrow functions (D3).
- **Function name (custos diverges):** upstream matches the written name
  `compact` case-sensitively and without resolution, so `Compact('typo')`
  is missed while a namespaced user function named `compact` (which does
  not read local variables) is checked. custos compares the name
  case-insensitively and requires the call to reach the builtin (D1).
