---
id: IssetArgumentExistence
group: Probable bugs
kind: semantic
needs: [flow]
php: { min: "", max: "" }
---

# IssetArgumentExistence

## Summary
`isset($v)`, `empty($v)` and `$v ?? …` silently accept a variable that is
never defined in the current function. When the first mention of a plain
variable inside a function body is such an existence check, the variable can
never be set at that point — usually a typo or a leftover after a refactoring.

## Detection
Only *plain* variables with a static name are examined (`$name`; not `$$n`,
`${expr}`, `$a['k']`, `$o->p`, `A::$p`).

Candidate positions:
- **D1** The left operand of a `??` binary expression, when that operand is
  directly a variable (no parentheses: `($v) ?? 1` is not a candidate). Only
  the outermost left operand of each `??` node is considered; in
  `$a ?? $b ?? $c` (right-associative: `$a ?? ($b ?? $c)`) only `$a` is a
  candidate (the right operand `$b ?? $c` is itself a `??` node whose left
  operand `$b` is also a candidate).
- **D2** Every argument of `isset(...)` that is directly a variable.
- **D3** The argument(s) of `empty(...)` that are directly a variable.

For each candidate variable `V` named `n`:
- **D4** `n` is not empty and is not one of the special names: `this`,
  `_GET`, `_POST`, `_SESSION`, `_REQUEST`, `_FILES`, `_COOKIE`, `_ENV`,
  `_SERVER`, `GLOBALS`, `HTTP_RAW_POST_DATA`, `php_errormsg`,
  `http_response_header`.
- **D5** `n` is not the name of a parameter of the innermost enclosing
  function-like `F` (function, method or closure), and not the name of a
  variable imported by `F`'s `use (...)` list (by-value or by-reference).
- **D6** `F` exists (the candidate is not in global/file scope) and `F` has a
  `{ }` body (arrow functions `fn() =>` have none → never reported).
- **D7** Walk every variable node of `F`'s own scope in source order (depth
  first, including `global`, `catch (T $e)`, `foreach (... as $k => $v)`,
  `list()`/`[]` destructuring, string interpolation, etc.) and take the
  **first** one named `n` — call it `R`. The walk stays in `F`'s scope:
  - a nested closure contributes only the variables of its `use (...)` list
    (they read or bind `F`'s variable); its parameters and body are skipped;
  - a nested arrow function contributes its body (it captures `F`'s
    variables), unless one of its parameters is named `n`;
  - nested named functions and class declarations contribute nothing.
  Report when either:
  - `R` is `V` itself, or
  - `R`'s direct parent is an assignment expression `A` (any assignment,
    including compound `.=`/`+=`… and by-reference `=&`) and the lowest common
    ancestor of `R` and `V` is exactly `A` — i.e. `V` lies inside that same
    assignment, as in `$n = $n ?? 0;` or `$n = isset($n);`.
  Otherwise (the first mention is something else) nothing is reported for `V`.
- **D8** Loop guard: walk upwards from `R`'s parent until reaching `F`; if a
  loop (`for`, `foreach`, `while`, `do … while`) is met, only the **first**
  (innermost) loop matters: suppress the report when anywhere inside that loop
  (any depth) there is an assignment whose left-hand side is directly a plain
  variable named `n` (this includes the assignment of D7's second case).
- **D9** Suppress when `F`'s body contains (any depth) a `goto` statement.
- **D10** Suppress when option `IGNORE_INCLUDES` is `false` and `F`'s body
  contains (any depth) an `include`, `include_once`, `require` or
  `require_once` expression.
- **D11** (custos, see Divergences) Suppress when `F`'s body defines
  variables dynamically: it calls the global `extract()`, the global
  `parse_str()` with a single argument, assigns a variable-variable
  (`$$name = …`, `${expr} = …`), or contains an include/require placed before
  `V` in the source (regardless of `IGNORE_INCLUDES`).

Each candidate is checked independently; several candidates of the same name
can each be reported (e.g. `isset($q) && empty($q)` as first two mentions:
only the first is the first mention, so only it is reported).

## Exceptions (no report)
- **E1** Any earlier mention of the variable in the function's own scope (a
  read, `echo $n`, `global $n`, a `use ($n)` of a nested closure, a read in a
  nested arrow function, a `catch` variable…). Mentions inside a nested
  closure body or named function do not count.
- **E2** Parameters and closure `use` imports of the enclosing function.
- **E3** Superglobals, `$this`, `$php_errormsg`, `$http_response_header`.
- **E4** File/global scope and arrow-function bodies.
- **E5** Variable assigned somewhere in the innermost loop that encloses the
  first mention (value carried over from a previous iteration).
- **E6** Functions containing `goto`.
- **E7** Functions containing include/require, only when
  `IGNORE_INCLUDES = false`.
- **E8** Non-plain arguments: `isset($a['k'])`, `isset($o->p)`, `empty(${$x})`.

## Report
- Range: the candidate variable node `V` (from `$` to the end of its name).
- Severity: **error** (fixtures tag it `error`, although the catalogue default
  is warning — upstream forces error-style highlighting).
- Message: `Variable '${name}' is not defined in this scope.`

## Fix
None.

## Options
| Option | Type | Default | Effect |
|---|---|---|---|
| `IGNORE_INCLUDES` | bool | `true` | When `true`, include/require statements in the function are disregarded (reports still happen). When `false`, a function containing any include/require is skipped entirely (the included file could define the variable). The EA fixture runs with `false`. |

## PHP versions
`??` exists from PHP 7.0. The EA fixture runs at PHP 7.1; isset/empty
detection has no version dependency.

## Examples

```php
<?php
class Billing
{
    public function total($items)
    {
        $sum = <error descr="Variable '$sum' is not defined in this scope.">$sum</error> ?? 0;
        $hit = isset(<error descr="Variable '$cache' is not defined in this scope.">$cache</error>, $items);
        $nil = empty(<error descr="Variable '$nil' is not defined in this scope.">$nil</error>);
        $cb  = function ($rate) use ($items) {
            return [$rate ?? 1, isset($items), $_POST ?? null, isset($this)];
        };
        return [$sum, $hit, $nil, $cb];
    }

    public function known()
    {
        echo $label;
        return $label ?? 'none';
    }

    public function looping($rows)
    {
        foreach ($rows as $row) {
            $prev = isset($last) ? $last : null;
            $last = $row;
        }
    }

    public function jumping()
    {
        again:
        $r = isset($tries) ? $tries : 0;
        $tries = 1;
        goto again;
    }

    public function including()
    {
        include 'defaults.php';
        return isset($config); // reported upstream when IGNORE_INCLUDES = true; never by custos (D11)
    }

    public function arrow()
    {
        return fn() => $missing ?? 0;
    }
}
$top = $undefinedAtTop ?? 1;
```

With `IGNORE_INCLUDES = false` the `including()` case is not reported (as
annotated above).

## Divergences
- `static $n;` declarations: whether upstream sees the declared name as a
  plain variable mention is unverified. Recommendation: treat `static $n`
  as an earlier mention (no report afterwards).
- **Scope-bounded first mention (custos diverges):** upstream's first-mention
  scan descends into nested closures and functions, so a variable that only
  exists inside an inner closure (`function () { $total = 0; … }`) hides a
  genuinely undefined `$total ?? …` in the outer function. custos stops the
  scan at scope boundaries (D7): only a closure's `use` list and arrow
  function bodies (which capture outer variables) count.
- **custos diverges — dynamically defined variables (D11).** Upstream
  reports `extract($config); … isset($host)` (Laravel's database
  connectors), `$$name = 1; isset($port)`, and, with the default
  `IGNORE_INCLUDES = true`, `include 'version.php'; … isset($wp_version)`
  (WordPress's update screens and loaders) as undefined variables at error
  severity, although those statements define them. custos suppresses the
  report when the function calls `extract()` or one-argument `parse_str()`,
  assigns a variable-variable, or includes a file before the check;
  `IGNORE_INCLUDES = false` still skips functions with any include.
- **Destructuring in the loop (custos diverges, D8).** A destructuring
  assignment listing the variable (`[$ts, $tz] = …`, `list(, list($x)) =
  …`) inside the innermost enclosing loop defines it for the next
  iteration, like a plain assignment (lazy initialisation in Craft's
  formatter).
- **Every enclosing loop (custos diverges from D8).** The loop guard checks
  every loop enclosing the first mention up to `F`, not only the innermost:
  `foreach ($ids as $id) { foreach (… as $k) { if (isset($prev)) … }
  $prev = $id; }` sets `$prev` for the outer loop's next iteration (tcpdf).
