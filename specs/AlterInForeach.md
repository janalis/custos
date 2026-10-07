---
id: AlterInForeach
group: Performance
kind: syntax
needs: []
php: { min: "", max: "" }
---

# AlterInForeach

## Summary
Three related `foreach` value-variable pitfalls:
1. a by-reference value (`as &$v`) that stays alive after the loop, so a later
   write to `$v` silently modifies the last array element;
2. an `unset($v)` right after a loop whose `$v` was *not* a reference (the
   unset is pointless, probably a leftover);
3. (opt-in) writing back through `$array[$key] = …` inside
   `foreach ($array as $key => $value)` where a by-reference value would do.

## Detection
Only foreach loops whose value target is a plain variable (`$v` or `&$v`)
are considered for parts A and B (value written as `list(...)`/`[...]`:
skip). "Next sibling" below means the next statement/node following a node in
its parent, skipping whitespace and ordinary comments (`//`, `#`, `/* */`) but
**not** doc comments (`/** */`), unless stated otherwise.

### A. By-reference value not unset
- **D1** The foreach value variable `V` is preceded by `&` (whitespace allowed
  between `&` and `$v`: `& $v`).
- **D2** Find the *follower* `N`:
  - `N` = next sibling of the foreach statement.
  - An `elseif`/`else` clause is never a follower: when the next sibling
    found at any step is such a clause, treat it as "none" (control leaves
    the branch for the statement after the whole `if`).
  - If there is none, climb ancestors starting from the foreach's parent:
    for each ancestor `P` that is **not** a `{ }` statement list (block), set
    `N` = next sibling of `P` (same clause rule; stop as soon as `N` is
    found). Then move to `P`'s parent; stop climbing when that parent is a
    function, method or closure (that body boundary is not crossed) or the
    file root.
    Consequences:
    - foreach as the last statement of a function/method/closure body or of
      the file → no follower;
    - last statement in any branch of an `if` chain (then, `elseif`, `else`,
      braced or braceless) → follower is the statement after the whole `if`
      statement;
    - last statement of a loop body → statement after that loop.
  - Then skip doc comments: while `N` is a doc comment, `N` = its next sibling.
- **D3** No report when any of these holds:
  - `N` does not exist;
  - `N` is a `return` statement or a `throw` statement;
  - `N` is an `unset(...)` statement with at least one argument that is a
    plain variable whose name equals `V`'s name (case-sensitive; other
    arguments ignored).
  Otherwise report `V` (severity warning).

### B. Unset of a non-reference value
- **D4** `V` is not preceded by `&`.
- **D5** Find `N` = next sibling of the foreach (doc comments are **not**
  skipped here, and no general ancestor climbing). While `N` is not an
  `unset` statement **and** the current foreach is the direct child of a
  `{ }` block that is itself the body of another foreach statement, move to
  that outer foreach and set `N` = its next sibling. (This climbs even when
  the inner foreach does have a non-unset follower inside the outer body.)
- **D6** `N` is an `unset(...)` statement. Each argument is tested on its
  own: every argument that is a plain variable whose name equals `V`'s name
  is reported (severity info), whatever the other arguments are
  (`unset($cache['x'], $v)` reports `$v`). Non-variable arguments
  (`$obj->v`, `$a['v']`) are never reported.

### C. Write-back through key (option `SUGGEST_USING_VALUE_BY_REF`, default off)
- **D7** Any assignment expression (plain `=`, by-reference `= &`, and
  compound `+=`, `.=`, …) whose left side is an array access `C[I]` with a
  non-empty index (`$a[]` excluded) where `I` is a plain variable.
- **D8** Walk the ancestors of the assignment up to the file root; stop (no
  report) on reaching a function/method/closure boundary. For each ancestor
  foreach that has a key variable, a value variable and an array expression:
  if the key variable is equivalent to `I` (same variable name) **and** the
  foreach array expression is equivalent to `C` (same node kind and
  structurally identical ignoring whitespace; two variables are equivalent
  when their names match) → report `C[I]` (severity info) and stop. Not
  matching → continue with outer ancestors (an outer foreach may match).
  When the matched foreach's value is already by reference (`as $k => &$v`),
  stop without reporting.
- **Name case.** Wherever this rule compares two expressions for
  equivalence, the names PHP resolves case-insensitively — function and
  method names, class names in calls, `new`, `instanceof` and `::`
  accesses, and keywords — compare case-insensitively (`Cache::$map['k']`
  matches `cache::$map['k']`, `$o->Name()` matches `$o->name()`); variable,
  property and constant names stay case-sensitive.

## Exceptions (no report)
- **E1** By-ref loop followed (directly or via the D2 climb) by
  `unset($v, ...)`, `return`, `throw`, or by nothing.
- **E2** Comments/doc comments between the loop and the unset/return
  (`/* */` always skipped; doc comments skipped in part A).
- **E3** Unset not naming the non-reference value variable as a plain
  variable argument (`unset($obj->v)` does not count).
- **E5** (custos, see Divergences) Part B: the loop is at file scope (no
  enclosing function, method or closure).
- **E4** Part C: option off; append form `$a[] = …`; non-variable index
  (`$a[$k + 1]`, `$a['x']`); container/key not matching the foreach; the
  assignment inside a closure/function nested in the loop.

## Report
- A: range = the value variable `V` only (`$item`, without `&`); severity
  warning. Message: `Unset '$item' right after the loop: it is still a
  reference to the last element.`
- B: range = the matching unset argument (`$item`); severity info (weak
  warning). Message: `'$item' is not a reference here; unsetting it is
  unnecessary.`
- C: range = the left-hand array access `C[I]` (e.g. `$rows[$idx]`); severity
  info. Message: `Iterate '$row' by reference and assign to it directly
  instead of writing through the key.` (`$row` = the foreach value variable
  name.)

## Fix
None.

## Options
| Option | Type | Default | Effect |
|---|---|---|---|
| SUGGEST_USING_VALUE_BY_REF | bool | false | Enables part C (D7–D8). Parts A and B are always active. |

## PHP versions
No gating. Upstream fixtures run at the test default level.

## Examples

```php
<?php
function adjust(array $prices, array $tags) {
    foreach ($prices as $sku => $amount) {
        <weak_warning descr="Iterate '$amount' by reference and assign to it directly instead of writing through the key.">$prices[$sku]</weak_warning> = $amount * 2;
    }
    unset($sku, <weak_warning descr="'$amount' is not a reference here; unsetting it is unnecessary.">$amount</weak_warning>);

    foreach ($tags as &<warning descr="Unset '$tag' right after the loop: it is still a reference to the last element.">$tag</warning>) {
        foreach ($prices as $p) {
            echo $tag, $p;
        }
    }
    unset(<weak_warning descr="'$p' is not a reference here; unsetting it is unnecessary.">$p</weak_warning>);

    foreach ($tags as &$t) { $t = trim($t); }
    /** trimmed */
    unset($t);

    if ($tags) {
        foreach ($tags as & $u) { $u .= '!'; }
    }
    return $tags;
}

function tail(array $list) {
    foreach ($list as &$entry) { $entry++; }
}
```

(Highlighting above assumes `SUGGEST_USING_VALUE_BY_REF = true`.)

## Divergences
- **Unset arguments tested individually (custos diverges):** upstream applies
  the "is a plain variable" test to the first unset argument instead of the
  argument being examined, so `unset($a['k'], $v)` reports nothing, while
  `unset($x, $obj->v)` after `foreach (... as $v)` reports the property
  fetch. custos tests each argument on its own (D6): the first case is
  reported, the second is not.
- **Branch followers (custos diverges):** upstream takes the next `elseif`/
  `else` clause as the follower of a by-ref loop that ends an `elseif` (or a
  braceless then-) branch, so the loop is reported even when `unset($v)`
  directly follows the whole `if`. That clause never runs after the branch,
  so custos looks past the `if` chain (D2) and stays silent in that case.
- **Part C and existing references (custos diverges):** upstream still
  suggests iterating by reference in `foreach ($a as $k => &$v) { $a[$k] =
  0; }`, where the value already is a reference; the advice is meaningless
  there, so custos does not report it (D8).
- Destructuring values (`as [$a, $b]`): upstream behaviour unverified;
  recommendation: skip parts A/B. Part C requires a value variable, so skip
  it too.
- **Name case (custos diverges).** Upstream compares the expressions
  textually, so operands that differ only in the case of a function, method
  or class name (`Stats::$n` vs `stats::$n`), which PHP treats as the same,
  are not recognised as equivalent. custos folds the case of those names
  (Detection, "Name case").
- **custos diverges — part B at file scope (E5).** Upstream reports
  `unset($item)` after a by-value loop everywhere. At file scope the loop
  variable is a global: the unset keeps it from leaking into the global
  scope (and into files included later, which WordPress's bootstrap files
  rely on — 17 findings). custos only reports part B inside functions,
  methods and closures.
