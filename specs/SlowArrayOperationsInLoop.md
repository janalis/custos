---
id: SlowArrayOperationsInLoop
group: Performance
kind: syntax
needs: []
php: { min: "", max: "" }
---

# SlowArrayOperationsInLoop

## Summary

Two loop anti-patterns:

- **Accumulating merges**: `$acc = array_merge($acc, $chunk)` inside a loop
  copies the whole accumulator on every iteration (quadratic cost). Collect
  the chunks and merge once after the loop.
- **Length in a `for` condition**: `$i < count($items)` re-evaluates the
  length on every iteration. Compute it once in the initialiser.

## Detection

### Part G — accumulating merges

- **G1** A function call (not a method/static call) that resolves to the
  global function (names compared case-insensitively, as PHP does; `\` and a
  global `use function` import are fine, but a same-named function declared in
  the current namespace, one imported from another namespace, or a qualified
  non-global name such as `Ns\array_merge` does not count): one of
  `array_merge`, `array_merge_recursive`, `array_replace`,
  `array_replace_recursive`. The message names the function in lower case.
- **G2** The call has **at least 2** arguments, and its 1st argument is not an
  array element access (`$a[$k]`, `$a['x'][1]`).
- **G3** The call's **direct** parent is an assignment expression `A` (plain
  `=` or compound such as `+=`) and the call is its value. Let `C` be the
  assignment target.
- **G4** Run-once filters, applied when `A` is an expression statement.
  Walk the ancestors of that statement, from its parent up to (excluding)
  the nearest enclosing loop, stopping at a function boundary:
  - **G4a** if an ancestor is a `{ … }` block whose last statement is a
    `break`, a `return` or a `throw` statement → no report (the merge runs at
    most once per loop);
  - **G4b** if an ancestor is an `if`, `elseif`/`else if` or `else` clause,
    or a `switch` case → no report (the merge is conditional).
  `try`/`catch`/`finally` and plain nested braces are transparent: the walk
  continues through them. When `A` is not a statement (part of a chained
  assignment, inside a `for` header, …), these filters are skipped.
- **G5** Walking up the ancestors of `A`'s statement (or of `A`'s parent when
  `A` is not a statement), stopping at the file or at the nearest enclosing
  function/method/closure/arrow function, at least one ancestor is a loop
  (`foreach`, `for`, `while`, `do … while`).
- **G6** `C` is equivalent to at least one of the call's arguments (any
  position, including the 1st). Equivalence: same node kind and, for simple
  variables, same name; for other expressions, structurally equal or
  identical source text (`$this->items` ≡ `$this->items`,
  `self::$cache` ≡ `self::$cache`).
- **G7** (custos, see Divergences) With `L` the innermost loop of G5, no
  report when the merge does not accumulate across iterations:
  - `C`'s base variable (`$v` of `$v`, `$v[…]`) is assigned by `L`'s header
    (foreach key/value, `for` initialiser/step) or by another plain `=`
    assignment inside `L` (a fresh value each iteration);
  - an index of `C` mentions such a variable (`$form[$id]`: another element
    each iteration);
  - `C` (for a plain variable: any mention of it) is read elsewhere in `L`,
    including the merge's other arguments and `L`'s condition
    (`$ctx = array_merge($ctx, $cb($ctx))`, `while ($g = array_shift($q))
    { $q = array_merge($q, …); }`): the next iteration needs the merged
    value.
  Other accumulating merges into the same `C` inside `L` are ignored by these
  checks.
- Report kind G on the call (once).

### Part F — length call in a `for` condition

- **F-1** A `for` statement. For each expression of its condition part (the
  middle, comma-separated section) that is a binary comparison `B`: one of
  `<`, `<=`, `>`, `>=`, `==`, `!=`/`<>`, `===`, `!==` (logical, arithmetic
  and other operators, including `<=>`, do not qualify):
- **F-2** Each **direct** operand of `B` (left and right; no parentheses in
  between) that is a function call (not a method/static call) resolving (as
  in G1) to one of the global `count`, `sizeof` (its alias), `strlen`,
  `mb_strlen` yields one report of kind F on `B`.
  (Arguments are not inspected.)
- **Name case.** Wherever this rule compares two expressions for
  equivalence, the names PHP resolves case-insensitively — function and
  method names, class names in calls, `new`, `instanceof` and `::`
  accesses, and keywords — compare case-insensitively (`Cache::$map['k']`
  matches `cache::$map['k']`, `$o->Name()` matches `$o->name()`); variable,
  property and constant names stay case-sensitive.

## Exceptions (no report)

- **E1** Merge call whose 1st argument is an element access.
- **E2** Merge result not directly assigned (returned, passed on, wrapped in
  parentheses), or the target not among the arguments.
- **E3** Merge nested (through any blocks, `try`/`catch`/`finally`) in a
  block ending with `break`/`return`/`throw`, or under an
  `if`/`elseif`/`else` clause or a `switch` case, within the loop (G4).
- **E4** Merge outside any loop of the current function (loops in an outer
  function do not count for code in a nested closure).
- **E5** Length already cached in the initialiser
  (`for ($i = 0, $n = count($a); $i < $n; $i++)`), empty condition, length
  call nested deeper than a direct operand (`$i < count($a) - 1`,
  `$i < (count($a))`), method calls (`$i < $c->count()`).
- **E6** (custos) Length of a value the loop body or step visibly writes
  (see Divergences): `for ($i = 0; $i < count($a); $i++) { $a[] = 0; }`.

## Report

- Range:
  - G: the merge call (name including qualifier through `)`).
  - F: the whole binary condition expression `B` (left operand start to right
    operand end).
- Severity: error for both.
- Messages:
  - G: `'{name}(...)' inside a loop re-copies the accumulator each time; merge once after the loop.`
  - F: `'{name}(...)' is re-evaluated on every iteration; compute it once before the loop.`

## Fix

Kind G: none.

- **F1 (kind F)** Let `L` be `B`'s length call and `O` the other operand. The
  call side is decided by whether `B`'s left operand is a call of any kind
  (see Divergences); in valid reports this is the side holding `L`.
  - Preferred name `{b}`: if `O` is a simple variable `$x`, `{b}` = `xMax`;
    otherwise (property, array access, call, literal, …) `{b}` = `loopsMax`.
  - **F1a** Variable name `{v}`: the candidates are `{b}`, `{b}1`, `{b}2`, …
    in that order; a candidate is *taken* when a variable of that name
    occurs anywhere in the enclosing scope (the nearest enclosing
    function/method/closure body including its parameters and anything
    nested in it, or the whole file for top-level code). Let `k` be the
    number of fixes with the same preferred name `{b}` that belong to `for`
    loops enclosing this one in the same scope, plus earlier conditions of
    this same `for`. `{v}` = `$` + the `(k+1)`-th candidate that is not
    taken. So an existing `$iMax` makes the fix use `$iMax1`, and two nested
    loops both preferring `loopsMax` get `$loopsMax` and `$loopsMax1`.
  - Replace `L` in `B` with `{v}` (the rest of `B` is untouched).
  - Insert the initialiser `{v} = {L}` (`{L}` = verbatim original call text):
    - if the `for` has at least one initial expression: insert `, ` +
      initialiser immediately after the last initial expression
      (`for ($i = 0; …` → `for ($i = 0, $iMax = count($a); …`);
    - if the initial section is empty: insert the initialiser immediately
      after the `(` of the `for` (`for (; …` → `for ($iMax = count($a); …`).
  - Examples:
    - `for ($i = 0; $i < count($rows); $i++)` →
      `for ($i = 0, $iMax = count($rows); $i < $iMax; $i++)`
    - `for ($k = 1; strlen($s) > $this->pos; $k++)` →
      `for ($k = 1, $loopsMax = strlen($s); $loopsMax > $this->pos; $k++)`

## Options

None.

## PHP versions

No gating. The EA tests run at the PhpStorm default level (between 5.6 and 7.0).

## Examples

```php
<?php
function collect(array $batches, $repo)
{
    $all = [];
    $map = [];
    foreach ($batches as $batch) {
        $all = <error descr="'array_merge(...)' inside a loop re-copies the accumulator each time; merge once after the loop.">array_merge($all, $batch)</error>;
        $map = <error descr="'array_replace(...)' inside a loop re-copies the accumulator each time; merge once after the loop.">\array_replace($batch, $map)</error>;
        try {
            $repo->rows = <error descr="'array_merge_recursive(...)' inside a loop re-copies the accumulator each time; merge once after the loop.">array_merge_recursive($repo->rows, $batch)</error>;
        } finally {
        }
    }
    while ($batch = array_shift($batches)) {
        $map[$batch['id']] = array_merge($map[$batch['id']], $batch);
        if ($batch) {
            $all = array_merge($all, $batch);
        }
        $other = array_merge($all, $batch);
    }
    foreach ($batches as $batch) {
        $all = array_merge($all, $batch);
        break;
    }
    $all = array_merge($all, $map);

    for ($n = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$n < count($batches)</error>; $n++) {}
    for ($n = 0, $m = 1; <error descr="'strlen(...)' is re-evaluated on every iteration; compute it once before the loop.">strlen($repo->name) >= $n</error>; $n++) {}
    for ($repo->at = 0; <error descr="'mb_strlen(...)' is re-evaluated on every iteration; compute it once before the loop.">$repo->at <= mb_strlen($repo->name)</error>; $repo->at++) {}
    for (; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">count($batches) > $n</error>; $n++) {}
    for ($n = 0, $total = count($batches); $n < $total; $n++) {}
    for ($n = 0; $n < count($batches) - 1; $n++) {}
    for ($n = 0; ; $n++) {}
    return [$all, $map];
}
```

```php
<?php
function collect(array $batches, $repo)
{
    $all = [];
    $map = [];
    foreach ($batches as $batch) {
        $all = array_merge($all, $batch);
        $map = \array_replace($batch, $map);
        try {
            $repo->rows = array_merge_recursive($repo->rows, $batch);
        } finally {
        }
    }
    while ($batch = array_shift($batches)) {
        $map[$batch['id']] = array_merge($map[$batch['id']], $batch);
        if ($batch) {
            $all = array_merge($all, $batch);
        }
        $other = array_merge($all, $batch);
    }
    foreach ($batches as $batch) {
        $all = array_merge($all, $batch);
        break;
    }
    $all = array_merge($all, $map);

    for ($n = 0, $nMax = count($batches); $n < $nMax; $n++) {}
    for ($n = 0, $m = 1, $nMax = strlen($repo->name); $nMax >= $n; $n++) {}
    for ($repo->at = 0, $loopsMax = mb_strlen($repo->name); $repo->at <= $loopsMax; $repo->at++) {}
    for ($nMax = count($batches); $nMax > $n; $n++) {}
    for ($n = 0, $total = count($batches); $n < $total; $n++) {}
    for ($n = 0; $n < count($batches) - 1; $n++) {}
    for ($n = 0; ; $n++) {}
    return [$all, $map];
}
```

## Divergences

- **Fix side detection (upstream bug):** upstream decides which operand is the
  length call by testing whether the left operand is *any* call, including
  method calls. For `$o->limit() > count($a)` it would replace the method
  call instead of `count(...)`. Recommendation: pick the side that holds the
  reported length call. Not covered by fixtures.
- **Both operands are length calls** (`count($a) < count($b)`): two reports
  on the same range with conflicting fixes (order unspecified upstream).
  Recommendation: report both, but offer the fix only for the left one.
- **Name clashes — custos diverges from upstream** (F1a). Upstream always
  uses `$xMax`/`$loopsMax`, overwriting a variable of that name that the
  code already uses (a parameter, a value read after the loop, the limit of
  an enclosing loop that got the same fix). custos picks the first numbered
  variant that does not occur in the scope and keeps nested loops apart. An
  EA fixture where the preferred name already occurs in the file now gets
  the numbered name; that case is listed in `testdata/ea-divergences.json`.
- **Comparisons only — custos diverges from upstream** (F-1). Upstream
  accepts any binary operator, so `$ok && count($a)` or `$i + strlen($s)` is
  reported as a loop limit and rewritten. Such conditions are not length
  comparisons; custos only considers relational and equality operators.
- **`sizeof` instead of `size` (custos diverges from upstream).** Upstream's
  length list contains `size`, which is not a PHP function (so any user
  function `size()` is treated as a length call), and misses `sizeof`, the
  built-in alias of `count`. custos lists `sizeof` and drops `size` (F-2).
- **Run-once filters through nested blocks — custos diverges from upstream**
  (G4). Upstream checks only the merge's immediate braced block: a merge in a
  `switch` case, in a `try` inside an `if`, in nested braces under an `if`,
  in a brace-less `if` body, or followed by `throw` is reported although it
  is conditional or runs once. custos walks up to the loop through nested
  blocks, treats `switch` cases as conditional, and counts `throw` as
  leaving the loop.
- **Function-name matching — custos diverges from upstream.** Upstream matches
  the merge and length functions by the name as written (case-sensitive, any
  namespace qualifier, no resolution), so a differently cased call such as
  `Array_Merge($all, $batch)` is missed while a namespaced or imported user
  function of the same name is reported (and rewritten) as if it were the
  builtin. custos matches case-insensitively and only calls that reach the
  global function.
- **Name case (custos diverges).** Upstream compares the expressions
  textually, so operands that differ only in the case of a function, method
  or class name (`Stats::$n` vs `stats::$n`), which PHP treats as the same,
  are not recognised as equivalent. custos folds the case of those names
  (Detection, "Name case").
- **custos diverges — merges that do not accumulate (G7).** Upstream reports
  every `C = array_merge(…, C, …)` in a loop at error severity, including
  `foreach ($items as $args) { $args = array_merge($defaults, $args); }`
  (a fresh value per iteration), `$form[$id] = array_merge([…],
  $form[$id])` (a different element per iteration) and accumulators the loop
  reads back (`$ctx = array_merge($ctx, $cb($ctx))`, work lists driven by
  `array_shift()`), none of which can be merged once after the loop (found
  on WordPress, Laravel and Drupal). custos skips them (G7).
- **Changing loop subjects (custos diverges, F1, E6).** The length is
  hoisted into the initialiser only when the measured value cannot change
  while the loop runs. When the body or step visibly changes it — assigns,
  pushes to, increments or unsets the measured variable or property, one of
  its elements or an offset of the measured element, or passes one of them
  to a by-reference parameter of a resolved function (`array_push()`,
  `array_pop()`, `sscanf()`) — the loop measures a length it changes on
  purpose (`$i < count($this->n)` with `$this->n['x'.$i] = 1;` in the
  body), and the condition is not reported at all (E6). When the change is
  only possible — the subject is not a variable or property fetch, or the
  body iterates it by reference, binds a reference to it, passes it to an
  unresolved callee or unpacks it, or calls a method on the object holding a
  measured property — the finding is reported without a fix
  (`$this->remove($i)` may shrink `$this->items`).
