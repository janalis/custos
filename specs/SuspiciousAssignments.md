---
id: SuspiciousAssignments
group: Probable bugs
kind: semantic
needs: [flow, types, names, index, hierarchy, stubs]
php: { min: "", max: "" }
---

# SuspiciousAssignments

## Summary

Groups six independent checks for assignments that are most likely mistakes:
writes lost to a switch fall-through, compound operators that repeat their
target (`$n += $n + 1`), parameters overwritten before ever being read,
`=+`/`=-`/`=!` typos, values written twice in a row, and array destructuring
of something that is not an array.

## Definitions

- **Plain assignment**: an assignment whose operator token is `=` (by value
  `$a = x` or by reference `$a = &x` / `$a =& x`). Compound assignments
  (`+=`, `.=`, `??=`, …) and destructuring (`list(...) =`, `[...] =`) are not
  plain assignments.
- **Target**: the left-hand side of an assignment.
- **Equivalent** (`≡`): two nodes are equivalent when they are of the same
  node kind and
  - for two simple variables: same name (`$a` ≡ `$a`; `$a` ≢ `$b`);
  - otherwise: same token sequence ignoring whitespace and comments, or
    identical source text.
  Parentheses are significant (`($a)` is a parenthesized node, not a variable).
- **Expression statement**: a statement consisting of one expression followed
  by `;`.
- **Statement list**: the direct statements of a `{ }` block, a function body,
  a `case`/`default` body or the file top level.

## Detection

### A. Switch fall-through overwrite

For every `switch`, walk its `case`/`default` clauses in order keeping a set
`W` of targets written by earlier clauses (initially empty).

- **D1** Clauses with no statements are skipped entirely (they neither add to
  nor clear `W`).
- **D2** For each direct statement of the clause, in order (statements nested
  in blocks, `if`, loops… are not looked at):
  - **D2a** destructuring expression statement (`list($a, $b) = …;` or
    `[$a, $b] = …;`): for each target variable of the list, if it is `≡` to an
    element of `W` → report that variable; otherwise remember it in the
    clause-local set `L`.
  - **D2b** plain-assignment expression statement with target `T`:
    - skip if `T` is an array append `X[] = …` (empty brackets);
    - skip if any node inside the right-hand side is `≡` to `T` (self-dependent
      write, e.g. `$f = sprintf($f, …)`);
    - if `T ≡` an element of `W` → report `T`; otherwise add `T` to `L`.
  - Other statements (calls, compound assignments `$y .= …`, echo…) are ignored.
  - Comparison is only against `W` (earlier clauses), never against earlier
    statements of the same clause.
- **D3** After the clause, add every element of `L` to `W` (deduplicated by `≡`).
- **D4** If the clause's last statement is `break`, `return`, `continue`,
  `goto`, `throw`, or an expression statement consisting of `exit`/`die`,
  clear `W`.

### B. Self-referencing compound assignment

- **D5** A compound assignment with operator one of `+= -= *= /= %= .= &= |=
  ^= <<= >>=` (not `**=`, `??=`).
- **D6** Its right-hand side is directly (not parenthesized) a binary
  expression whose operator is the matching binary operator (`+` for `+=`,
  `.` for `.=`, …), and whose **left operand** is `≡` to the target. Binary
  operators are left-associative, so in `$s .= $s . 'a' . 'b'` the left operand
  of the outer `.` is `$s . 'a'`, which is not `≡ $s` → no report.

### C. Parameter overwritten before use

Applies to every function, method and closure (not arrow functions) with a
body, unless it is in a test context (E6).

- **D7** The function has at least one parameter and its body has at least
  one statement.
- **D8** For each parameter not passed by reference, list all accesses (reads
  and writes) of a variable with that name in the function's control flow, in
  evaluation order starting at the function entry (right-hand sides are
  evaluated before the write of an assignment; code inside nested closures is
  not part of the flow). Require at least 2 accesses.
- **D9** The first access is the target variable of a plain assignment, and
  that assignment is an expression statement placed **directly** in the
  function body (not nested in `if`, loops, `try`, another expression…).
- **D10** Within that whole assignment expression (target and right-hand side,
  including nested closures and strings), the parameter's name occurs exactly
  once (as the target).
- Report the target variable.

### D. `=+`, `=-`, `=!` formatting typo

- **D11** A plain assignment (anywhere) whose right-hand side is a unary
  expression with operator `+`, `-` or `!`, where
  - the `=` token is immediately followed by the unary operator (no whitespace
    or comment between them), and
  - the unary operator is immediately followed by whitespace.
  `$a =- $b` matches; `$a = -$b`, `$a=-$b`, `$a =-$b` do not.

### E. Value overwritten immediately

- **D12** A plain assignment that is itself an expression statement; target `T`.
- **D13** `T` is not an "unpredictable array write": walking down through
  array-access targets (`X[k]` → `X`), stop and skip the check if any level has
  - empty brackets (`X[]`), or
  - a key (parentheses stripped) that is the constant `__LINE__`, or
  - a key that is neither a string literal nor a number literal (negative
    number literals count as numbers) — e.g. a variable or a call.
  Only `T`s whose array keys are all string/number literals (or that are not
  array accesses at all) continue.
- **D14** `T` is not used by its own assignment: no node of the right-hand
  side, the right-hand side itself included, is `≡ T` (so `$v = $v;` is
  skipped);
  and no variable name other than `this` occurs twice among the variables
  strictly inside `T` (e.g. `$m[count($m)]`).
- **D15** Let `P` be the previous statement in the same statement list
  (comments skipped). No `P` → nothing.
- **D16** If `P` is an `if` statement:
  - it has no `elseif`/`else` branch;
  - its body is a braced block (an unbraced body → nothing);
  - the block's last statement is not `return`, `continue`, `break`, `throw`
    or an `exit`/`die` expression statement (an empty block → nothing);
  - find the first direct statement of the block that is a **plain**
    assignment expression statement whose target is `≡ T` (compound
    assignments such as `$v .= 'x'` do not count); none → nothing;
  - let `S` be the statement following it inside the block; if `S` is an `if`,
    consider only its condition, otherwise the whole statement. If any node
    strictly inside that region is `≡ T` (same kind) → nothing;
  - otherwise report the **whole assignment expression** (case "conditional").
- **D17** Otherwise, if `P` is an expression statement whose expression is a
  plain assignment with target `≡ T`:
  - skip when `P` assigns by reference (`= &…`);
  - skip when `P` sits directly in the block of a `try` (not `catch`/`finally`);
  - skip when `T` contains a `++` or `--` operator (prefix or postfix);
  - otherwise report `T` only (case "general").

### F. Destructuring a non-array

- **D18** A destructuring assignment (`list(...) = v` or `[...] = v`) that is
  itself an expression statement (not in `foreach`, not nested).
- **D19** Resolve the type set of `v`, dropping unknown parts. If the set is
  empty → nothing.
- **D20** Report when at least one type in the set does not support
  destructuring. Supporting types: `array` (including any `X[]` typed array),
  `mixed`, and classes/interfaces whose inheritance tree (itself, parents,
  implemented interfaces) contains `\ArrayAccess`. Everything else — `string`,
  `int`, `float`, `bool`/`true`/`false`, `null`, `callable`/`\Closure`,
  `iterable`, `object`, `resource`, `self`/`static`/`$this`, classes without
  `ArrayAccess`, `void` — is non-supporting.
- **D20a** (custos refinement, see Divergences) Failure markers: when the set
  contains at least one supporting type, `null` and `false` are ignored for
  D20 (they are the usual "no result" members of an array-returning call:
  `?array`, `array|false`, `int[]|false|null`). A set made only of `null`
  and/or `false`, or one where another non-supporting type remains
  (`array|string`, `int[]|int|false`), is still reported.
- **Name case.** Wherever this rule compares two expressions for
  equivalence, the names PHP resolves case-insensitively — function and
  method names, class names in calls, `new`, `instanceof` and `::`
  accesses, and keywords — compare case-insensitively (`Cache::$map['k']`
  matches `cache::$map['k']`, `$o->Name()` matches `$o->name()`); variable,
  property and constant names stay case-sensitive.

## Exceptions (no report)

- **E1** Switch: writes in clauses separated by a terminating statement (D4);
  self-dependent writes; appends `X[] = …`; writes in nested blocks; empty
  clauses in between do not break the chain.
- **E2** Compound assignment whose RHS is parenthesized, whose RHS's left
  operand is parenthesized (`$n += ($n) + 1`) or uses a different operator.
- **E3** Parameter cases: by-reference parameters; first access is a read
  (`$p = trim($p)`); first write is conditional/nested; the parameter is
  never accessed again after the write; the name appears again in the
  assignment (e.g. in a closure `use`).
- **E4** Formatting: whitespace between `=` and the operator, or no
  whitespace after the operator.
- **E5** Sequential writes: append/dynamic-key targets, `__LINE__` keys,
  self-referencing RHS (`$v = trim($v)`, `"{$v}"`), preceding `if` with
  `else`/`elseif` or terminating last statement or a consumer right after the
  conditional write; preceding write by reference; preceding write directly in
  a `try` block; `++`/`--` in the target; preceding compound assignment
  (`$v .= 'x'; $v = '';`), also as the only write inside the preceding `if`
  (`if ($c) { $v .= 'x'; } $v = '';`); a self-assignment `$v = 1; $v = $v;`.
- **E6** Test context (check C only): file path ends with `Test.php`,
  `Spec.php` or `.phpt`, or contains `/Fixtures/`; or the enclosing class FQN
  ends with `Test` or contains `\Tests\` or `\Test\`.
- **E7** Destructuring of `array`, `mixed`, `ArrayAccess` implementations, or a
  value whose type is entirely unknown.
- **E8** Supporting type plus only `null`/`false` failure markers (D20a):
  `[$k, $v] = $maybeRow` with `array|false`, `?array`, `array|null|false`.

## Report

Severity: error for all checks.

| Check | Range | Message |
|---|---|---|
| A | the overwritten target (variable, property, array access, or list element) | `This write overwrites a value set in a previous case; a 'break' may be missing.` |
| B | the whole compound assignment expression (target through RHS, no `;`) | `The target is repeated on the right-hand side of the compound assignment; likely a merge mistake.` |
| C | the target variable of the overwriting assignment | `Parameter is overwritten before its value is used.` |
| D | the whole assignment expression (target through RHS, no `;`) | `Did you mean '{op}'? Fix the operator or the spacing.` where `{op}` is `+=`, `-=` or `!=` |
| E conditional | the whole assignment expression (no `;`) | `{target} is overwritten right after the 'if'; an 'else' may be missing.` |
| E general | the target only | `{target} is overwritten right after being assigned.` |
| F | the whole destructuring expression (no `;`) | `Destructuring a value that is not an array.` |

`{target}` is the target's source text.

When several checks apply to the same assignment they are reported
independently (D and E both run on plain assignments).

## Fix

None.

## Options

None.

## PHP versions

No gating. Upstream fixtures run at the test-default level (below 7.1); the
destructuring fixture uses `[...] =` and a `mixed` parameter type anyway, so
parse these regardless of level.

## Examples

```php
<?php
function classify($kind) {
    switch ($kind) {
        case 'x':
            $label = 'ex';
        case 'y':
            <error descr="This write overwrites a value set in a previous case; a 'break' may be missing.">$label</error> = 'why';
            break;
        case 'p':
            $pair = 1;
        case 'q':
            [<error descr="This write overwrites a value set in a previous case; a 'break' may be missing.">$pair</error>, $rest] = [2, 3];
            return $pair;
        case 'r':
        case 's':
            $acc[] = $kind;
            $msg = 'm';
            $msg = strtoupper($msg);
            break;
    }
}

<error descr="The target is repeated on the right-hand side of the compound assignment; likely a merge mistake.">$total -= $total - 1</error>;
<error descr="The target is repeated on the right-hand side of the compound assignment; likely a merge mistake.">$path .= $path . '/'</error>;
$path .= $path . '/' . 'x';
$total *= ($total * 3);

function normalize($name, $size, &$out) {
    $name = strtolower($name);
    <error descr="Parameter is overwritten before its value is used.">$size</error> = strlen($name);
    $out = [];
    return [$name, $size];
}

<error descr="Did you mean '-='? Fix the operator or the spacing.">$delta =- $step</error>;
<error descr="Did you mean '!='? Fix the operator or the spacing.">$ok =! $failed</error>;
$delta = -$step;
$other=+$step;

function flow($cond, $list) {
    if ($cond) { $mode = 'a'; }
    <error descr="$mode is overwritten right after the 'if'; an 'else' may be missing.">$mode = 'b'</error>;

    if ($cond) { $cfg["k"] = 1; return; }
    $cfg["k"] = 2;

    $count = 0;
    <error descr="$count is overwritten right after being assigned.">$count</error> = count($list);

    $list[] = 1;
    $list[] = 2;
    $tmp = 'x';
    $tmp = "<{$tmp}>";
    return [$mode, $cfg, $count, $tmp];
}

function unpack_all(int $n, array $row, \ArrayObject $obj) {
    [$a, $b] = $row;
    list($c, $d) = $obj;
    <error descr="Destructuring a value that is not an array.">[$e, $f] = $n</error>;
    <error descr="Destructuring a value that is not an array.">list($g) = new \DateTime()</error>;
}

/** @return int[]|false */
function sample_pair() { return [1, 2]; }

function failure_markers(array|false $found, ?array $cached, string|false $line, false $none) {
    [$p, $q] = $found;                   // E8: false ignored next to array
    [$r] = $cached;                      // E8: null ignored next to array
    [$s, $t] = sample_pair();            // E8: int[]|false
    <error descr="Destructuring a value that is not an array.">[$u] = $line</error>;   // string remains
    <error descr="Destructuring a value that is not an array.">[$w] = $none</error>;   // only false
}
```

## Divergences

- Check C relies on control-flow ordering; if the implementation lacks a CFG,
  approximate with a source-order scan of the body where an assignment's RHS
  precedes its target, and nested closures are excluded from the scan (their
  text still counts for D10).
- **D16 — custos diverges from upstream.** Upstream takes any assignment in
  the preceding `if` block as the conditional write, so
  `if ($c) { $v .= 'x'; } $v = '';` is reported with "an 'else' may be
  missing". A compound assignment updates the existing value rather than
  choosing an alternative one, so an `else` is not the fix the message
  suggests, and D17 already exempts a preceding compound write. custos counts
  only plain `=` writes in the `if` block. No upstream fixture covers it.
- **D14 — custos diverges from upstream.** Upstream compares only nodes
  strictly inside the right-hand side, so `$v = 1; $v = $v;` is reported as
  an overwrite although the second statement keeps the value. custos also
  compares the right-hand side itself and stays silent.
- **D20a — custos refinement, not upstream.** Upstream reports any
  non-supporting member, so a nullable array (`?array`, `array|null`) or an
  array-or-false result (`array|false`, typical of built-ins returning
  `false` on failure, e.g. `int[]|false` from stubs) is reported although the
  value is an array whenever destructuring can succeed (found on real code).
  custos ignores `null` and `false` when at least one supporting type is
  present. The upstream fixtures report `string` and `stdClass` values only,
  so conformance is unaffected. Note: a stub union that also carries other
  scalars (e.g. `array|int|float|false`) is still reported; silencing such
  calls needs argument-aware return types, not this rule.
- Check A for property/array targets uses the same equivalence, so
  `$this->a` written in two fall-through clauses is reported too. Keep.
- **Name case (custos diverges).** Upstream compares the expressions
  textually, so operands that differ only in the case of a function, method
  or class name (`Stats::$n` vs `stats::$n`), which PHP treats as the same,
  are not recognised as equivalent. custos folds the case of those names
  (Detection, "Name case").
- **D16 reads and exits (custos diverges).** Upstream looks for a read of
  `T` only in the statement following the conditional write (only its
  condition when it is an `if`), and only checks whether the block's last
  statement is a jump. So `if ($warm) { $r = get($k, $hit); if ($hit) {
  return $r; } } $r = build();` and a block ending in an `if/else` whose
  branches both return are reported, although the conditional value is used
  or the code after the `if` is never reached from it. custos scans every
  later statement of the block, skips blocks that always terminate
  (`syntax.Terminates`), and — when `T` is shared state (see below) — also
  skips when a later statement of the block makes a call, which may read it.
- **D17 shared targets (custos diverges).** `global $r; $r = load(); $r =
  tidy();`, `$_SESSION['k'] = []; $_SESSION['k'] = merge();` or
  `$this->p = 1; $this->p = $this->compute();` are reported by upstream,
  but the call on the second right-hand side can read the first value. When
  `T` is a property, static property, variable variable, superglobal,
  file-scope variable, `global`/`static` variable, by-reference parameter or
  a variable bound by reference, custos skips D16/D17 if the relevant code
  contains a call (function, method, static call, `new` or include).
- **Reads through the holding array (custos diverges).** For an element
  target such as `$package['version']`, a read of the array holding it as a
  whole (`$loader->load($package)`, `$q = $package['a']` for
  `$package['a']['b']`) counts as a read of the value in D14 and in the
  later statements of D16 (Composer `PathRepository`: the conditional value
  was passed to `load($package)` before being replaced). Reads of sibling
  elements (`$package['name']`) do not count.
- **`true`/`bool` as failure markers (custos diverges, extends D20a).**
  Next to a supporting type, `true` and `bool` are ignored like `null` and
  `false`: `@return array|bool` guarded by `if (!$r) return;` leaves
  `array|true`, a loose documentation of "array or false" (Magento price
  filter). `bool` alone is still reported.
- **Check A reads before the overwrite (custos diverges).** A fall-through
  that reads the earlier case's value before writing it again
  (`case 1: $errors = check(); if ($errors !== []) { return 1; } // no
  break` then `case 2: $errors = more();`) loses nothing, so it is not the
  missing-`break` bug. A target read by a statement after its write (in its
  own case or in a later one, before the overwrite; for an assignment, its
  right-hand side) is dropped from `W` (Mautic installer steps).
- **Check C positional reads (custos diverges).** A function that calls
  `func_get_arg()` or `func_get_args()` (outside nested functions) reads
  the passed values without naming the parameter, so its parameters are
  not reported (`$b = func_num_args() > 1 ? func_get_arg(1) : null;`
  keeps a deprecated signature working).
