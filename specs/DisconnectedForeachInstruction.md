---
id: DisconnectedForeachInstruction
group: Control flow
kind: semantic
needs: [names, index, types, stubs]
php: { min: "", max: "" }
---

# DisconnectedForeachInstruction

## Summary

A statement inside a `foreach` body that uses no loop variable and nothing the
loop body changes produces the same effect on every iteration; it most likely
belongs before or after the loop. Optionally (off by default) the rule also
points at objects instantiated on every iteration that could be built once and
cloned.

## Detection

The rule runs on every `foreach` statement `L`.

### Preconditions

- **D1** `L` has a braced body (`{ … }`, a group statement). Bodies without
  braces (`foreach ($a as $b) stmt;`) are not analysed.
- **D2** If any direct child of the body is inline HTML (the body contains a
  `?> … <?php` section), the whole loop is skipped.
- **D3** The *statements* analysed are the direct children of the body that are
  real statements/elements; comments and docblocks are ignored.

### Loop variables (seed of the "modified" set `M`)

- **D4** Walk from `L` upwards through its ancestors (including `L` itself),
  stopping at the nearest enclosing function, method, closure or arrow
  function, or the file. For every `foreach` met on the way, add to `M` the
  names of its key variable, value variable, and every variable inside a
  `list(…)` / `[…]` destructuring used as its value. `for`/`while` loops do
  not contribute.

### Pass 1 — per statement dependencies `Dep(S)` and global writes `M`

All statements are scanned in pass 1 before any is judged, so a variable
written by a later statement still counts as "modified" for an earlier one.

For each statement `S`, visit every variable node `$v` (name `n`) anywhere
inside `S` (deep, including nested closures, string interpolation, catch
variables, inner loop headers). Let `C` be `$v`, extended outwards while its
parent is a property access whose object is `C` (`$v->a->b` → `C` is
`$v->a->b`). Let `P` be the parent of `C`. If `P` is an element of an array
literal, replace `P` by the array literal, and if that literal is itself the
child of something, by that something (so `[$a, $b] = …` and `f([$a])` see the
assignment / argument list). Rules, applied in order (first that ends with
"stop" wins):

- **D5** `P` is a destructuring assignment (`list(…) = …` or `[…] = …`) and
  `$v` is not its right-hand side: add `n` to `M` and `Dep(S)`; stop.
- **D6** `P` is any other assignment (plain `=`, by-reference, or compound
  `+=`, `.=`, …) whose left-hand side is exactly `C`: add `n` to `M`. Also add
  `n` to `Dep(S)` when the assignment is compound, or when `C` is a property
  access (`$v->p = …`), or when the assignment itself is a direct call
  argument (`f($v = new X())`). Stop.
- **D6a** `$v` is the caught variable of a `catch (T $v)` clause, or lies in
  the key or value target of a `foreach` nested in `S` (including
  destructuring targets `as [$a, $b]`): record `n` as **bound by `S`**;
  stop. A name bound by `S` makes *other* statements that depend on it
  connected (the value changes on every iteration), but does not connect
  `S` itself (see pass 2).
- **D7** `P` is an array access whose base is `C` (`$v[...]`, `$v->p[...]`):
  add `n` to `Dep(S)`. Add `n` to `M` only when the element is written:
  climb from `P` through further array/property accesses on it to the
  outermost access `O`; it is a write when `O` is the left-hand side of an
  assignment (any operator), the right-hand side of a by-reference
  assignment (`$r = &$v['k']`), the operand of `++`/`--`, an argument of
  `unset(…)`, the object of a method call (`$v['q']->push(…)`), or a direct
  argument of a resolved function at a by-reference parameter
  (`sort($v['k'])`). Plain reads (`echo $v['k']`) do not modify. Do not
  stop (later rules can only add `n` to `Dep(S)` again).
- **D8** `P` is a call argument list and `$v` is directly an argument:
  - **D8a** the call is an instance method call with `->` (not `::`, not
    `?->`): take its object expression and strip parentheses, property
    accesses (to their object) and array accesses (to their base). If that
    yields a variable `$o`, add `$o`'s name to `M` and stop **without** adding
    `n` to `Dep(S)` (the argument is considered absorbed by the object).
    Otherwise continue.
  - **D8b** the call is a plain function call that resolves to a known
    function (user-defined or stub) whose parameter at the same position as
    `$v` is declared by-reference: add `n` to `M` and `Dep(S)`; stop.
    Unresolved functions or by-value parameters: continue.
- **D8c** (custos, see Divergences) `$v` (or a property chain rooted at it)
  is the receiver of an instance method call that is itself an expression
  statement (its result is discarded: `$bar->advance();`): add `n` to `M`
  and `Dep(s)`.
- **D9** `P` is a `++`/`--` operation (prefix or postfix): add `n` to `M` and
  `Dep(S)`; stop.
- **D10** Otherwise add `n` to `Dep(S)`.
- **D11** Additionally, for every call to the global function `compact`
  (name compared case-insensitively; unqualified, `\compact`, not a
  namespaced or shadowing user function of that name) anywhere inside `S`, each argument that is a string literal with non-empty
  contents adds those contents (the variable name without `$`) to `Dep(S)`.

### Pass 2 — judging each statement

`S` is **connected** when `Dep(S)` contains a name that is in `M`, or that
is bound (D6a) by a statement other than `S`. Connected statements are never
reported. For an unconnected statement, classify it (looking through an
expression statement `expr;` to `expr`):

| class | shape |
|---|---|
| control | `break`, `continue`, `return` statements |
| increment / decrement | `++x` / `x++` / `--x` / `x--` |
| new | `$var = new …` (left side is a plain variable) |
| clone | `$var = clone …` |
| dom-create | `$var = <expr>->createElement(…)` where the call resolves to `DOMDocument::createElement` |
| assignment | any other assignment whose left side is a plain variable (incl. compound) |
| accumulate | assignment whose left side is an array push without index (`$a[] = …`, `$o->p[] = …`) |
| other | everything else (calls, `echo`, `unset`, `if`, loops, `switch`, `try`, assignments to `$a[k]` or `$o->p`, …) |

- **D12** (disconnected) Class is *other*, the statement is a real statement
  (not a bare expression fragment such as the content of `<?= ?>`), has
  non-zero length, contains strictly inside it **no** `break`, `continue`,
  `return` or `throw` (searched deep, including nested closures), and contains
  at least one variable node strictly inside it → report.
- **D13** (clone suggestion) Only when option `SUGGEST_USING_CLONE` is on:
  class is *new* or *dom-create* → report the whole statement with the clone
  message. (Independent of D12, which never applies to those classes.)

## Exceptions (no report)

- **E1** Statements depending on a loop variable (current or any enclosing
  `foreach` up to the function boundary), on any variable written anywhere
  in the body, or on a catch / inner-foreach variable bound by another
  statement (D4–D11).
- **E2** Assignments to plain variables, `$x[] = …` pushes, `++`/`--`,
  `break`/`continue`/`return` — never reported as disconnected.
- **E3** Statements containing `break`, `continue`, `return` or `throw`
  anywhere inside (e.g. `if ($limitReached) { break; }`).
- **E4** Statements without any variable (`echo '<hr>';`, `flush();`).
- **E5** Bodies containing inline HTML; loops without braces.
- **E6** `$obj->m($loopVar)` style mutations: the object becomes "modified",
  so later `$obj->other()` statements are connected (D8a).
- **E7** Variables written through by-reference parameters of resolved
  functions (`preg_match(…, $m)`, `array_pop($stack)`) (D8b).
- **E8** `compact('name')` referencing a loop/modified variable (D11).

## Report

- Range (D12):
  - when the statement is `if`, `while`, `do … while`, `for`, `foreach`,
    `switch` or `try`: only its leading keyword token (`if`, `while`, `do`,
    `for`, `foreach`, `switch`, `try`);
  - any other statement: the whole statement including its trailing `;`.
- Range (D13): the whole statement including the trailing `;`.
- Severity: info (weak warning) for both.
- Messages: D12 `Statement does not depend on the loop; move it out.`;
  D13 `Create the object once before the loop and clone it here.`

## Fix

None.

## Options

| Option | Type | Default | Effect |
|---|---|---|---|
| SUGGEST_USING_CLONE | bool | false | Enables D13 (object creation on every iteration). The upstream test suite runs with it enabled. |

## PHP versions

No gating. Short list syntax `[$a, $b] = …` (7.1+) follows D5 the same way as
`list(…)`.

## Examples

```php
<?php
/* @var array $orders */
$type = \ArrayObject::class;
foreach ($orders as $key => $order) {
    <weak_warning descr="Create the object once before the loop and clone it here.">$box = new $type();</weak_warning>
    <weak_warning descr="Create the object once before the loop and clone it here.">$node = (new \DOMDocument())->createElement('row');</weak_warning>

    <weak_warning descr="Statement does not depend on the loop; move it out.">syslog(LOG_INFO, $banner);</weak_warning>
    <weak_warning descr="Statement does not depend on the loop; move it out.">while</weak_warning> ($pending > 0) {
        usleep($pending);
    }
    <weak_warning descr="Statement does not depend on the loop; move it out.">if</weak_warning> ($debugMode) {
        dump($settings);
    }
    <weak_warning descr="Statement does not depend on the loop; move it out.">switch</weak_warning> ($mode) {
        default:
    }
    <weak_warning descr="Statement does not depend on the loop; move it out.">try</weak_warning> {
        $warmup();
    } catch (\RuntimeException $problem) {
        report($problem);
    }

    echo '<br>';                        // E4: no variables
    if ($stop) { return; }              // E3
    $total += $order;                   // E2 + connected
    $copy = clone $template;            // clone: never reported
    $sink[] = 'seen';                   // accumulate
    $logger->push($order);              // D8a: $logger becomes modified
    $logger->flush();                   // connected through $logger
    preg_match('/x/', $order, $hits);   // D8b: $hits modified
    print_r($hits);                     // connected
    $meta->touch(compact('key'));       // D11: depends on $key
}

foreach ($groups as $group) {
    foreach ($group as $member) {
        echo $group;                    // D4: outer loop variable
    }
}

foreach ($rows as $row) {
    ?><td><?= $title ?></td><?php       // E5: inline HTML, loop skipped
}
```

## Divergences

- Throw: upstream searches for a `throw` strictly *inside* the statement, so a
  bare `throw $error;` statement directly in the body may be reported as
  disconnected depending on how the parser nests the throw (older parsers:
  statement is the throw itself → reported; newer: expression statement
  containing a throw expression → not reported). Recommendation: never report a
  statement that is or contains `throw`. No fixture covers the bare form.
- **Element reads (custos diverges):** upstream marks an array as modified
  on any element access, so a pure read such as `echo $cfg['title'];` makes
  every statement using `$cfg` connected and hides genuine findings. custos
  only counts element writes (D7).
- D8a ignores nullsafe `?->` calls. Recommendation: treat `?->` like `->`.
  No fixture coverage.
- **Catch and inner-loop variables (custos diverges):** upstream treats them
  as plain reads, so a statement reading the variable after the inner loop
  or `try` (`report($line);` after `foreach ($batch as $line)`) is reported
  although its value changes on every iteration. custos treats them as
  bindings of the declaring statement (D6a): other statements using them are
  connected, while the declaring statement itself is still judged on its own
  dependencies.
- **compact() matching (custos diverges):** upstream recognises only the
  exact spelling `compact` regardless of namespace, so `\COMPACT('key')`
  (which reads `$key`) left the statement reported as disconnected, while a
  user function `compact` declared in the current namespace (which cannot
  read the caller's variables) hid a genuine finding. custos matches the
  name case-insensitively and only when the call resolves to the global
  built-in (D11).
- **custos diverges — method calls made for their side effect (D8c).**
  Upstream reports `$bar->advance();`, `$stack->pop();`,
  `$writer->endElement();` in a loop as not depending on it, although each
  call changes the object on every iteration (about 13 of 18 sampled
  findings on Drupal, Laravel and Nextcloud). custos treats the receiver of
  a method call whose result is discarded as modified. Calls whose result
  is used (`echo $o->label();`) are unchanged.
- **Per-iteration effects (custos diverges).** Statements calling a
  stream-writing, random or clock built-in (`fwrite`, `fputs`, `fputcsv`,
  `fprintf`, `vfprintf`, `fflush`, `file_put_contents`, `rand`, `mt_rand`,
  `random_int`, `random_bytes`, `lcg_value`, `uniqid`, `microtime`,
  `hrtime`, `time`, `array_rand`, `shuffle`, `str_shuffle`) are not
  reported: running them once before the loop changes the output or the
  values (PrestaShop writes one `.htaccess` header per shop; Matomo draws a
  random value per period). `error_log()` and `usleep()` keep the upstream
  behaviour (upstream fixtures). Output is per-iteration too: statements
  containing `echo` or `print`, or calling `printf`/`vprintf`/
  `var_dump`/`print_r`/`var_export`/`debug_zval_dump`/
  `debug_print_backtrace`, are not reported (Moodle prints
  `$OUTPUT->box_start()` per row, progress dots under `if ($display) {
  echo '.'; }`); listed upstream divergence.
- **Repeat loops (custos diverges).** A loop whose body never reads its
  own key or value variable (`foreach (range(1, 10) as $attempt) {
  post(…); }`, Bagisto's login-throttle test) exists to repeat the body:
  nothing in it is reported. Bodies calling `compact()`,
  `get_defined_vars()`, `extract()`, using variable variables, `include`
  or `eval` may read the variable and are checked as before.
- **By-reference arguments of methods (custos extends D8b).** D8b also
  applies to resolved instance, static and constructor calls:
  `StringHelper::stringIncrement($column)` (`string &$str`) changes
  `$column` on every iteration (PhpSpreadsheet). For an instance call the
  receiver is still marked modified (D8a).
- **Round-4 refinements (custos diverges).** Found on Shopware, Akeneo,
  Kimai and API Platform (14 of 20 findings):
  - D8c also applies at the root of a discarded method chain
    (`$qb->where(…)->setParameter('k', $row);`,
    `$ctx->getContext()->addState(…);`, `$this->getBrowser()->request(…);`),
    and D8's argument case strips method calls from the receiver like
    property fetches: the root variable counts as modified.
  - Once a statement of the body may leave the iteration without being a
    jump itself (a `continue`/`break` not bound by a nested loop or switch,
    `return`, `throw`, `exit`; nested functions skipped), every later
    statement is connected: it runs on some iterations only (`if ($i !==
    $w) { continue; } notify($w);`).
  - Two statements writing the same variable are connected
    (`$level = 1; if (isset($p['d'])) { $level = $p['d'] + 1; }`): moving
    either out of the loop changes what the other sees.
- **Callbacks (custos diverges).** A statement calling a callback held in
  a variable or other expression (`$progress && $progress(['type' =>
  'progress']);`, `$log('step');`) is not reported: like a progress tick
  it is meant to run once per item (Grav's archivers and security scans).
