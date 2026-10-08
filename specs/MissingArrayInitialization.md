---
id: MissingArrayInitialization
group: Probable bugs
kind: syntax
needs: []
php: { min: "", max: "" }
---

# MissingArrayInitialization

## Summary

Appending to a local array (`$list[] = …`) deep inside nested loops without
ever initialising it means the variable is undefined (`null`) when nothing is
appended, and its contents leak between outer iterations if it is meant to be
reset. The array should be initialised explicitly at the right place.

## Detection

- **D1** An array-push access: an array access expression whose index
  brackets are empty (`X[]`), in any context.
- **D2** The innermost enclosing function-like `F` (function, method, closure,
  arrow function) exists — code at file/global scope is never reported.
- **D3** Walking upwards from the access's parent to `F`, at least **two**
  loop statements (`for`, `foreach`, `while`, `do … while`, in any
  combination) are crossed.
- **D4** Strip all array accesses from `X` (`$v[$i][$j][]` → `$v`): the base
  must be a plain variable `$name` (not a property, static property, call,
  `$$dyn`).
- **D5** `F` has a `{ }` body (arrow functions have none → no report).
- **D6** `$name` is not a parameter of `F` and not imported by `F`'s
  `use (...)` list.
- **D7** Scan every plain variable named `$name` anywhere in `F`'s body (any
  depth, before or after the access, nested closures included). Do not report
  if any of them:
  - is a **direct operand of an assignment expression** — either side, any
    assignment operator (`=`, `=&`, `.=`, `+=`, `??=`, …): e.g.
    `$name = [];`, `$name = load();`, `$copy = $name;`;
  - is a **direct part of a `foreach` header** — the iterated expression or
    the key/value target: `foreach ($name as …)`,
    `foreach ($rows as $name)`, `foreach ($rows as $name => $v)`.
  Variables wrapped in another node (`$name[] = 1`, `$name['k'] = 1`,
  `count($name)`, `[$name] = …`) do not count.
- Otherwise report the access.

Each qualifying `X[]` access is reported independently.

## Exceptions (no report)

- **E1** Fewer than two enclosing loops inside the function.
- **E2** Global scope, arrow functions.
- **E3** Parameters and closure `use` imports.
- **E4** Any direct assignment to/from the variable or a foreach-header
  occurrence anywhere in the body.
- **E5** Non-variable bases: `$this->items[]`, `self::$cache[]`, `f()[]`.

## Report

- Range: the whole array-push access expression, from the start of the base
  variable to the closing `]` of the empty brackets (`$v[$i][]`). The
  assigned value is not included.
- Severity: warning.
- Message: `Array '${name}' is never initialised; initialise it before the loops.`

## Fix

None.

## Options

None.

## PHP versions

None.

## Examples

```php
<?php
function collect(array $groups, $limit) {
    for ($n = 0; $n < $limit; ++$n) {
        do {
            <warning descr="Array '$bucket' is never initialised; initialise it before the loops.">$bucket[]</warning> = $n;
            <warning descr="Array '$bucket' is never initialised; initialise it before the loops.">$bucket[$n][]</warning> = $limit;
        } while (false);
    }

    foreach ($groups as $group) {
        $flat[] = $group;
    }

    $done = [];
    foreach ($groups as $group) {
        foreach ($group as $item) {
            $done[] = $item;
            $limit[] = $item;
            $group[] = 0;
            $this->seen[] = $item;
        }
    }

    return function ($extra) use ($done) {
        while (true) {
            foreach ($extra as $e) {
                $done[] = $e;
                $extra[] = $e;
            }
        }
    };
}

foreach ($a as $x) {
    foreach ($x as $y) {
        $global[] = $y;
    }
}
```

## Divergences

- `global $name;` and `static $name = [];` declarations do not suppress the
  report upstream (only direct assignments/foreach headers do), which gives
  false positives. Recommendation: also treat `global`/`static` declarations
  of the name as initialisation.
- Whether a variable inside a `list()`/`[...]` destructuring target counts as
  an assignment operand upstream is unverified. Recommendation: treat it as
  an assignment (no report).
- **Superglobals (custos).** `$_SESSION['k'][] = $v`, `$GLOBALS['x'][] = $v`
  and the other superglobals hold state from outside the function; they
  are never reported (SuiteCRM's upgrade wizard).
