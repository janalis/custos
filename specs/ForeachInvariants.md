---
id: ForeachInvariants
group: Control flow
kind: semantic
needs: [names, index, stubs]
php: { min: "", max: "" }
---

# ForeachInvariants

## Summary
Two loop shapes that are really array iterations in disguise:
1. a counter-based `for` that walks `0 … count($a)` and reads `$a[$i]`;
2. a `while (list($k, $v) = each($a))` loop (`each()` is slow, deprecated in
   PHP 7.2 and removed in PHP 8).
Both are clearer (and faster) as `foreach`.

## Detection

### Part A — counter `for` loops
A `for` statement `L` is reported when all of the following hold:

- **D1** `L` has exactly one "step" expression (third clause).
- **D2** `L` has a braced body containing at least one statement (comments and
  docblocks don't count).
- **D3** Counter: among the init expressions (first clause, any position) the
  first plain `=` assignment (not compound `+=` etc.) whose left side is a plain variable and whose right side's source
  text is exactly `0` defines the counter `$i`. `0.0`, `00`, `-0`, `'0'` do not
  qualify. Other init expressions may precede or follow
  (`$n = count($a), $i = 0` is fine).
- **D4** The single step expression is `++$i` or `$i++` on that same variable.
- **D5** Limit: `L` has exactly one condition expression, it is a binary
  expression, and one of its operands is a plain variable with the counter's
  name; the *other* operand is the limit `X`. The operator itself is not
  checked (see Divergences).
- **D6** Container: collect every array access anywhere in the body (deep,
  including nested closures/loops) whose index is exactly the counter variable
  (`…[$i]`; not `[$i + 1]`). Their base expressions, compared by source text,
  must all be the same single expression `$c` (stop as soon as two distinct
  texts are found → no report). No such access → no report.
- **D7** The limit must stand for `count($c)`: compute the possible values of
  `X` (see D8); there must be exactly one, and it must be a plain function call
  to the global function `count` (name compared case-insensitively, as PHP
  does; a namespaced or shadowing user `count` does not count) with exactly
  one argument structurally
  equal to `$c` (ignoring whitespace). `sizeof`, `strlen`, method calls,
  `count($c, COUNT_RECURSIVE)` → no report.
- **D8** Possible-values discovery of an expression `E` (recursively, each
  node at most once):
  - parentheses are stripped;
  - ternary `a ? b : c` / `a ?: c` → values of the true and false branches;
  - `a ?? b` → values of both operands;
  - a variable → the default value of a same-named parameter of the enclosing
    function/method/closure (if any), plus the right-hand sides of **every**
    plain `=` assignment to that variable anywhere in that function's body
    (chains `$a = $b = v` contribute the innermost `v`) — except for the
    limit `X` itself, see D8a. Outside any function
    (top-level code) a variable has **no** values, so a variable limit is never
    accepted there. **Unstable variable** (custos refinement, see Divergences): if the variable is also the operand of `++`/`--` or the target of a compound assignment (`+=`, `.=`, `??=`, …) anywhere in that scope body, the whole result is *unknown* → no report;
  - a property access → property default value (unless the default text ends
    with the property name) plus plain assignments to the same property
    access in the enclosing method and in the class constructor;
  - class constant → its resolved value; global constant → its `define()`
    value (`true`/`false`/`null` are kept as-is);
  - anything else → itself.
- **D8a** When the limit `X` is a plain variable, only the definitions that
  reach the loop condition count: if `L`'s own init clause assigns `X` with
  plain `=` (`for ($i = 0, $n = count($a); …)`), that assignment alone
  (the last one in the clause) gives the value; otherwise the plain `=`
  assignments that reach the `for` statement (same reaching-definition
  approximation as `DynamicCallsToScopeIntrospection` R3: a later
  assignment statement in an enclosing statement list hides earlier ones;
  assignments after the loop count only through an enclosing loop), plus
  the parameter default when no dominating assignment precedes. The values
  of those right-hand sides are then discovered with D8 as usual. A limit
  variable reused by several loops is therefore accepted when each loop is
  preceded by its own `$n = count(…);`, while a limit that may hold two
  values at the loop (`$n = count($a); if ($f) { $n = count($b); }`) is
  still rejected.
- **D8b** Write guard (custos refinement, see Divergences): neither the
  counter `$i` nor the limit `X` (when `X` is a plain variable or a property
  access) may be written anywhere inside `L` other than in its standard
  places. The standard places are: the counter's own `$i = 0` init (D3), the
  single step expression (D4), and, for the limit only, one plain `=`
  assignment whose target is `X` among the init expressions
  (`$i = 0, $n = count($a)`). Any other write within the init expressions,
  the condition or the body (deep, including nested loops; closures only
  through by-reference `use (&$i)` imports, arrow functions never) counts:
  - assignment of any kind (`=`, compound `+=`, `.=`, `??=`, …, by-reference
    `= &`) whose target is the variable, or destructuring that lists it
    (`[$i, $x] = …`, `list(, $n) = …`);
  - `++` / `--` (prefix or postfix) on it;
  - passing it as an argument at a by-reference parameter position of a call
    that resolves to a known function/method (user code or stubs);
    unresolved calls are not considered writes;
  - `foreach (… as $i)` / `foreach (… as $k => $i)` value or key target,
    `unset($i)`, `global $i`, `static $i`, `&$i` in a by-reference
    assignment's right side (`$r = &$i`), or capture by reference in a
    closure `use (&$i)`.
  Property-access limits compare by structural equality; a method call on the
  property's object does not count as a write.
  Any such write → no report.
- **D8d** Container changes (custos refinement, see Divergences): no report
  when, in `L`'s init, condition or body, the container `$c` itself is
  assigned (any operator, destructuring, by-reference), passed to a
  by-reference parameter (`sort($c)`, `array_splice($c, $i, 1)`), unset,
  pushed to (`$c[] = …`), or one of its elements is unset. `foreach`
  iterates a snapshot of the array while the counter loop reads the live
  one, so such loops visit different elements. Element writes
  (`$c[$i] = …`) are allowed.
- **D8c** Use after the loop (custos refinement, see Divergences): no
  report when the counter `$i`, or a limit `X` that is a plain variable
  assigned in `L`'s own init clause, is mentioned anywhere after the end of
  `L` in the same scope (the enclosing function/method/closure body, or the
  whole file for top-level code; nested named functions and classes are
  other scopes). A `foreach` leaves different values behind (no counter, or
  the last key instead of the element count; no limit at all once the
  header is gone), so a later read would change meaning. Exception, without
  flow analysis: the first later mention does not count when it is the
  target of a plain `=` assignment (not by reference) whose value does not
  mention the variable, and that assignment is a statement of its own
  (`$i = …;`) or an init expression of a `for` header, in the statement list
  that holds `L` or one of `L`'s enclosing statements (so it dominates every
  later mention). `for ($i = 0; …) {…} for ($i = 0; …) {…}` is therefore still
  reported for both loops, while a re-assignment inside an `if`, a call
  argument, `$i = $i + 1` or `$i += 1` keeps the first loop unreported.
- **D9** Report on the `for` keyword (severity warning) and offer fix F1.

### Part B — `each()` loops
- **D10** A destructuring assignment (`list(…) = …` or `[…] = …`) whose
  right side (after unwrapping a bare expression wrapper) is a plain function
  call written unqualified that resolves to the global function `each`
  (case-insensitive name; not a user `each` declared or imported in the
  current namespace) with exactly one argument `$c`.
- **D11** The assignment is directly the condition of a `while` loop or
  directly a clause of a `for` loop, and that loop has a braced body with at
  least one statement.
- **D12** No node in the body (deep) of the same kind as `$c` is structurally
  equal to `$c` (for a variable: same name). If the body mentions `$c` at all,
  no report.
- **D13** Report on the loop keyword (`while` / `for`) with **error**
  severity. Fix F2 is offered only for `while` loops whose destructuring has
  exactly two target variables; otherwise the report has no fix.
- **Name case.** Wherever this rule compares two expressions for
  equivalence, the names PHP resolves case-insensitively — function and
  method names, class names in calls, `new`, `instanceof` and `::`
  accesses, and keywords — compare case-insensitively (`Cache::$map['k']`
  matches `cache::$map['k']`, `$o->Name()` matches `$o->name()`); variable,
  property and constant names stay case-sensitive.

## Exceptions (no report)
- **E1** More than one step expression, or a step other than `++` on the
  counter (`$i += 1`, `$i--`).
- **E2** Counter not initialised to literal `0`.
- **E3** Condition list with 0 or 2+ expressions, or a condition that is not
  a binary expression directly comparing the counter (`$i < $n || $stop`).
- **E4** Two different containers indexed by the counter (`$a[$i]`, `$b[$i]`).
- **E5** Limit not provably `count($c)` (parameter, literal, `strlen`,
  variable with several assignments reaching the loop, top-level variable).
- **E6** Empty or brace-less body.
- **E7** `each()` loops whose body uses the iterated container itself.
- **E8** Counter or limit written inside the loop outside its standard places
  (D8b): `$i += 2;` / `$i--` / `$n--` in the body, `$i = 0, $i = 1` in the
  init, `$i < ($n = count($a))` in the condition, `advance($i)` with a
  by-reference parameter, `foreach ($x as $i)` in the body.
- **E9** Counter, or header-assigned limit, used after the loop (D8c): a
  search loop `for (…) { if (…) break; } if ($i == count($a))`, or
  `echo $n;` after `for ($i = 0, $n = count($a); …)`.

## Report
- Range: the loop's first keyword token (`for` or `while`).
- Severity: Part A warning (rule default); Part B error.
- Messages: Part A `Iterate with foreach instead of a counter loop.`;
  Part B `Replace the each() loop with foreach.`

## Fix

### F1 — counter loop → foreach
Let `$i` be the counter text, `$c` the container text, and `$iValue` the
counter text with `Value` appended (`$i` → `$iValue`, `$idx` → `$idxValue`).

1. Inside the body, replace qualifying `$c[$i]` accesses (base structurally
   equal to `$c`, index structurally equal to `$i`) by `$iValue`. An access
   qualifies only when its direct parent is one of:
   - a property access or method call on it (`$c[$i]->p`, `$c[$i]->m()`,
     including as an assignment target: `$c[$i]->p = 1`);
   - a binary expression; a unary expression (`!`, `-`, casts, …);
     parentheses; use as an index of another access (`$x[$c[$i]]`);
   - an `echo` statement (as one of its arguments);
   - the condition of `if` / `elseif`, the subject of `switch`, a `case`
     value;
   - a loop header clause (`for` clauses, `while` / `do … while` condition,
     `foreach` source);
   - string interpolation: `"… $c[$i] …"` **and** `"… {$c[$i]} …"` both become
     `"… $iValue …"` (the braces are dropped), except that `{$iValue}` is
     written when the next character would extend the interpolation (an
     identifier character, `[` or `->`: `"{$c[$i]}abc"` → `"{$iValue}abc"`,
     `"$c[$i][0]"` → `"{$iValue}[0]"`);
   - an argument of a function or method call that resolves to a known
     function none of whose parameters (any position) is by-reference;
     unresolved calls → not replaced;
   - the right-hand side of a non-by-reference assignment (`$x = $c[$i]`,
     `$x .= $c[$i]`); by-reference (`$x = &$c[$i]`, `$x =& $c[$i]`) → kept;
   - the base of a further array access (`$c[$i][0]`), unless the outermost
     array access of that chain is directly inside an assignment (either side)
     → kept (`$c[$i][0] = 1` and `$y = $c[$i][0]` are kept; `echo $c[$i][0]`
     becomes `echo $iValue[0]`).
   Any other position (assignment target `$c[$i] = …`, `return $c[$i]`,
   array literal element, `isset`/`unset`, `new X($c[$i])`, ternary branch,
   `yield`, bare statement…) is left unchanged.
2. Replace the whole `for (…) <body>` with
   `foreach (<c> as <i> => <iValue>) <body>` (single spaces; body text as
   modified in step 1, kept verbatim).
3. If, after step 1, no variable named like the counter remains anywhere in
   the body, drop `<i> => ` so the header is `foreach (<c> as <iValue>)`.
4. Limit cleanup: in the enclosing function body, count nodes structurally
   equal to the limit `X` (after the loop was replaced). If exactly one
   remains and it is the left side of a plain assignment that is a statement
   on its own (`$n = count($c);`), delete that statement (with its line).
   Nothing is deleted at top level or when the limit was a direct
   `count(...)` call.

### F2 — each loop → foreach (while only)
With the two destructuring targets `$k`, `$v` and argument `$c`: replace the
whole loop with `foreach (<c> as <k> => <v>) <body>`; if the body has no
variable named like `$k`, drop `<k> => `. No body rewriting, no cleanup.

## Options
None.

## PHP versions
No gating (Part B is reported on every level, even where `each()` no longer
exists).

## Examples

```php
<?php
function render(array $rows, $cap) {
    $total = count($rows);
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($n = 0; $n < $total; $n++) {
        print $rows[$n];
        echo $rows[$n]->title, " ({$rows[$n]}) ";
        $label = strtoupper($rows[$n]);
        $alias = &$rows[$n];
        if ($rows[$n] instanceof Stringable) {}
    }
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($n = 0; count($rows) > $n; ++$n) {
        $rows[$n][1] = 'x';
        echo $rows[$n][2];
    }
    <error descr="Replace the each() loop with foreach.">while</error> (list($slot, $entry) = each($rows)) {
        echo $entry;
    }
    for ($n = 0; $n < $cap; $n++) { echo $rows[$n]; }               // E5: parameter
    for ($n = 1; $n < count($rows); $n++) { echo $rows[$n]; }       // E2
    while (list($slot, $entry) = each($rows)) { unset($rows[$slot]); } // E7
}
```

```php
<?php
function render(array $rows, $cap) {
    foreach ($rows as $n => $nValue) {
        print $rows[$n];
        echo $nValue->title, " ($nValue) ";
        $label = strtoupper($nValue);
        $alias = &$rows[$n];
        if ($nValue instanceof Stringable) {}
    }
    foreach ($rows as $n => $nValue) {
        $rows[$n][1] = 'x';
        echo $nValue[2];
    }
    foreach ($rows as $entry) {
        echo $entry;
    }
    for ($n = 0; $n < $cap; $n++) { echo $rows[$n]; }               // E5: parameter
    for ($n = 1; $n < count($rows); $n++) { echo $rows[$n]; }       // E2
    while (list($slot, $entry) = each($rows)) { unset($rows[$slot]); } // E7
}
```

```php
<?php
function bump(&$pos) { $pos++; }
function scan(array $tokens) {
    $limit = count($tokens);
    for ($k = 0; $k < $limit; $k++) {          // E8: counter skipped ahead
        if ($tokens[$k] === '(') { $k += 2; }
    }
    for ($k = 0; $k < $limit; $k++) {          // E8: limit shrinks
        if ($tokens[$k] === '') { $limit--; }
    }
    for ($k = 0; $k < $limit; ++$k) {          // E8: by-reference argument
        echo $tokens[$k];
        bump($k);
    }
    for ($k = 0; $k < $limit; ++$k) {          // E8: overwritten by a nested foreach
        echo $tokens[$k];
        foreach ([1, 2] as $k) {}
    }
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($k = 0, $size = count($tokens); $k < $size; ++$k) {
        echo $tokens[$k];                      // init assignment of the limit is standard
    }
}
```

Notes: `print` is not in the list of replaced contexts, so `print $rows[$n]`
stays and the key `$n =>` is kept (same for the by-reference assignment).
After the first loop is replaced, `$total` occurs only in its own assignment,
so step 4 deletes `$total = count($rows);`.

## Divergences
- **Unstable variables — custos refinement, not upstream.** Value discovery ignores `++`/`--` and compound assignments upstream, so a variable later incremented or extended is analysed with its initial value only. custos makes the result unknown (no report), as in the shared value discovery of `CallableMethodValidity`. No upstream fixture relies on such a variable; recorded in `docs/decisions.md` ("Spec-level false positives").
- D5 accepts any binary operator between counter and limit (`<=`, `>`, `!=`,
  even `+`), so `for ($i = 0; $i <= count($a); $i++)` is reported and fixed
  although it iterates one element too many. Recommendation: only accept
  `$i < X`, `X > $i`, `$i != X`, `X != $i` (and `!==`). No fixture covers
  other operators.
- **String interpolation (custos diverges).** Upstream turns
  `"{$c[$i]}abc"` into `"$iValueabc"` (another variable) and
  `"$c[$i][0]"` into `"$iValue[0]"` (an offset read instead of literal
  text). custos keeps braces (`{$iValue}`) when the next character is an
  identifier character, `[` or `->`, and drops them otherwise, as upstream
  does.
- `++`/`--` applied to `$c[$i]` are unary expressions upstream and get
  rewritten (`$c[$i]++` → `$iValue++`), silently losing the write. custos
  does not replace operands of `++`/`--` (they are not in the F1.1 list).
- **Reused limit variables (custos diverges):** upstream collects every
  assignment to the limit variable in the function, so a `$n` reused by
  several loops — each re-assigning it right before — has several values
  and none of the loops is reported. custos only considers the assignments
  that reach the loop (D8a).
- **Write guard (D8b/E8) — custos refinement, not upstream.** Upstream
  only checks the header shape, so a loop whose counter is advanced or reset
  in the body (`$i += 2`, `$i--`, `skip($i)` by reference) or whose limit
  changes (`--$n`, `$n = count($a)` inside the body) is still reported, and
  the suggested `foreach` silently changes the iteration (found on real code:
  tokenizers skipping ahead, loops shrinking their bound). custos skips any
  loop whose counter or limit is written anywhere except its standard init
  and step. No upstream fixture writes the counter or limit inside a
  reported loop (the only writes in reported loops are the limit's own init
  assignment, which stays allowed), so conformance is unaffected. Recorded in
  `docs/decisions.md` ("Spec-level false positives").
- **Use after the loop (D8c/E9) — custos refinement, not upstream.**
  Upstream ignores what happens after the loop, so a search loop whose
  counter is tested after a `break`, or a header limit read afterwards, is
  reported and the fix leaves the later code reading a stale or undefined
  variable. custos skips such loops (conservative syntactic check described
  in D8c, no flow analysis). The upstream fixture loses two reports
  (counters mentioned again after their loop without a dominating
  re-assignment); the case was already a listed divergence.
- Part B requires the iterated argument to be absent from the body; the fix
  does not check that the iterated array's internal pointer state is relied
  upon afterwards.
- **Function names (custos diverges):** upstream matches `count` (D7) and
  `each` (D10) by exact spelling regardless of namespace, so `COUNT($c)` /
  `Each($c)` loops were missed while loops over a same-named user function
  in the current namespace were reported (the suggested `foreach` would then
  drop that function's behaviour). custos matches the names
  case-insensitively and only when they resolve to the global built-ins.
- **Name case (custos diverges).** Upstream compares the expressions
  textually, so operands that differ only in the case of a function, method
  or class name (`Stats::$n` vs `stats::$n`), which PHP treats as the same,
  are not recognised as equivalent. custos folds the case of those names
  (Detection, "Name case").
- **Container changes (D8d) — custos refinement, not upstream.** Upstream
  rewrites `for ($i = 0; $i < $n; $i++) { … array_splice($posts, $i, 1); … }`
  (WordPress sticky-post reordering), loops that `sort()`, reassign, push to
  or unset elements of the iterated array; the `foreach` replacement walks a
  snapshot and changes behaviour. custos skips them. No upstream fixture
  changed outcome.
