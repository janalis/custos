---
id: UsingInclusionOnceReturnValue
group: Probable bugs
kind: syntax
needs: []
php: { min: "", max: "" }
---

# UsingInclusionOnceReturnValue

## Summary

`include_once`/`require_once` return the file's value only the first time;
every later evaluation returns `true`. Code that uses that value works once
and then silently breaks. Use `include`/`require` when the result matters.

## Detection

- **D1** An inclusion expression whose keyword is `include_once` or
  `require_once`, in any letter case (`REQUIRE_ONCE`, `Include_Once`; see
  Divergences), and that has an operand.
- **D2** The inclusion is **not** the whole expression of a plain expression
  statement — its direct parent is anything else: an assignment right-hand
  side, a `return`, an `if`/`while` condition, a parenthesized expression, a
  binary operand, a call argument, an array element, `@`, another inclusion's
  operand, etc.

Note the low precedence of inclusion: in `require_once $a && require_once $b`
the operand of the first inclusion is `$a && require_once $b`, so the outer
inclusion spans up to the end of `$b`. The inner one is a logical operand,
which is a success test (E3): only the outer inclusion can be reported.

## Exceptions (no report)

- **E1** `require_once 'x.php';` / `include_once('y.php');` as standalone
  statements.
- **E2** `include`/`require` (non-`_once`) anywhere.
- **E3** The result only says whether the file could be included (see
  Divergences), looking through parentheses and `@`:
  - *discarded*: `@include_once $f;`, `(@require_once $f);`;
  - *success test*: an `if`/`elseif`/`while`/`do-while` or ternary
    condition, an operand of `!`, `&&`, `||`, `and`, `or`, `xor`, a
    comparison with `false` (`===`, `!==`, `==`, `!=`), or a `(bool)` cast;
  - *success flag*: the right-hand side of a plain `=` to a simple variable,
    the assignment being a statement or itself a success test, when every
    read of that variable in the scope is a success test and the variable
    is never extended (`$v[] =`, `$v['k'] =`, `$v->p =`, `.=`, `++`).
    Inside a function a variable that is never read also qualifies, unless
    the scope contains another inclusion (the included file shares the
    scope's variables). At file scope a never-read variable is reported.

  A variable is **not** a success flag when other code can reach it:
  - it is declared `global` or `static` in the scope;
  - it is a by-reference parameter of the scope, or a by-reference `use` of
    the closure that is the scope;
  - the scope (nested arrow functions included; nested functions, closures
    and classes excluded) calls `compact()` or `get_defined_vars()`, uses a
    variable-variable (`$$name`, `${expr}`) or `eval`.

## Report

- Range: the inclusion expression, from the keyword to the end of its operand
  (the operand may span lines and contain other inclusions). Surrounding
  parentheses and the `;` are not included.
- Severity: error.
- Message: `Only the first include_once/require_once returns the file's value; later ones return true.`

## Fix

None. Dropping `_once` would re-run the file on every evaluation (see
Divergences).

## Options

None.

## PHP versions

No gating.

## Examples

```php
<?php
$routes = <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">require_once __DIR__ . '/routes.php'</error>;
$ok = <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once $first ||
    include_once $second</error>;
while (include_once($plugin)) {
    register((<error descr="Only the first include_once/require_once returns the file's value; later ones return true.">require_once $plugin</error>));
}
require_once __DIR__ . '/bootstrap.php';
include_once('helpers.php');
$cfg = (require 'config.php');

function loadTranslations(string $file)
{
    $loaded = @include_once $file;
    if (!$loaded) {
        return false;
    }
    return true;
}

function loadSettings(string $file)
{
    global $settings;
    $settings = <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once $file</error>;
}

function loadSettingsInto(string $file, array &$settings)
{
    $settings = <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">require_once $file</error>;
}

function renderPage(string $file)
{
    $page = <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once $file</error>;
    include __DIR__ . '/layout.php';
}
```

## Divergences

- **Keyword case (custos diverges).** Upstream requires the keyword text to
  end with lowercase `_once`, so `$cfg = REQUIRE_ONCE $f;` is not reported
  although PHP keywords are case-insensitive and the value is just as
  unreliable. custos matches the keyword in any case (D1).
- **Success tests (custos diverges).** `include_once`/`require_once`
  return `false` only when the file cannot be included, so a result that is
  only tested for success is reliable: custos does not report an inclusion
  (through parentheses and `@`) used as an `if`/`elseif`/`while`/`do-while`
  or ternary condition, as an operand of `!`, `&&`, `||`, `and`, `or`,
  `xor`, or compared with `false` (`===`, `!==`, `==`, `!=`) — e.g.
  `if (!include_once $lib)`, `(@include_once $f) !== false` (E3).
  Upstream reports them all.
- **No quick-fix (custos diverges).** Replacing `_once` with a plain
  `include`/`require` (upstream F1/F2) re-runs the file on every call, which
  redeclares the classes and functions it defines (fatal) and repeats its
  side effects. custos reports without a fix; whether to load the value
  another way is the author's decision.
- **Silenced statements (custos diverges).** `@include_once $f;` is an
  expression statement whose result is discarded through the `@`
  (PrestaShop); D2 now looks through `@` and parentheses.
- **Bool cast (custos diverges).** `(bool) include_once $path` (Joomla's
  loader: `$found = (bool) include_once $path; if ($found) …`) is a
  success test like `!include_once`: not reported. Other casts still are.
- **Success flags (custos diverges).** A result stored in a plain local
  variable (`$found = @include_once $dir . $f; if ($found) break;`,
  `$res = include_once $f; if (!$res) die();`) is a success flag when
  every read of the variable in its scope is a success test as above (and
  it is never extended with `[]`/`.=`): not reported. Inside a function a
  variable that is never read is not reported either; at file scope it
  is (the including file may read it). About 22 of 36 reports on Dolibarr.
  The exemption stops where the variable is not private to the scope
  (E3): a `global` or `static` variable, a by-reference parameter or
  closure use, a scope that reads variables by name (`compact()`,
  `get_defined_vars()`, `$$name`, `eval`) and, for a never-read variable,
  a scope with another inclusion. There the stored value is observable
  and turns into `true` on the second evaluation, so the report is kept.
