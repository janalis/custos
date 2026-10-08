---
id: AmbiguousMethodsCallsInArrayMapping
group: Performance
kind: syntax
needs: []
php: { min: "", max: "" }
---

# AmbiguousMethodsCallsInArrayMapping

## Summary

When a loop builds a map like `$map[key($x)] = key($x);`, the same call is
evaluated twice per iteration (once for the key, once for the value). Compute
it once into a local variable.

## Detection

- **D1** A `for` or `foreach` statement whose body is a statement block (the
  `{ ... }` block; the statement list of the alternative `: ... endfor;` /
  `endforeach;` syntax counts as a block too). A loop whose body is a single
  braceless statement is not analysed.
- **D2** Each **direct** statement of that block is examined (statements
  nested in `if`, inner blocks, inner loops, … are not examined from this
  loop; inner loops are analysed on their own). The statement must be an
  expression statement whose expression, after removing any enclosing
  parentheses (`($m[f()] = f());`), is an assignment (plain `=`, `= &`,
  or compound `.=`, `+=`, … ) with a right-hand side, and whose left side is
  an array access `C` (`X[...]`, at any nesting: `$a[..][..]`).
- **D3** `L` = all function calls and method calls (instance `->`, nullsafe
  `?->` and static `::`) that are descendants of `C` — anywhere in `C`,
  i.e. in the index, in nested indexes, or in the accessed base expression;
  `C` itself is never a call. If `L` is empty, stop.
- **D4** `R` = all function/method calls that are descendants of the
  right-hand side `V`, in source order (pre-order, outer before inner), and,
  if `V` itself is a call, `V` appended at the **end** of the list.
- **D5** Iterate `R` in that order, skipping every `r` that is not stable
  (E6); for the first remaining `r ∈ R` for which some
  `l ∈ L` has the same called name (function name or method name,
  compared case-insensitively as PHP resolves them — custos diverges, see
  Divergences) **and** `l` is equivalent to `r` (same node kind and
  structurally identical, ignoring whitespace and comments; plain variables
  compare by name; keywords, and function, method and class names in calls,
  `new`, `instanceof` and `::` accesses, compare case-insensitively at any
  depth, while variables, properties and constants stay case-sensitive) →
  report `r` and stop processing this statement (at most one report per
  statement).

`new` expressions, language constructs (`isset`, `empty`, `array(...)`,
`list`, `echo`, `print`, `include`) are not calls for this rule.

## Exceptions (no report)

- **E1** Call only on one side (`$m[f($x)] = 'k';`, `$m['k'] = f($x);`).
- **E2** Same name but different arguments/receiver (`$m[f($a)] = f($b);`,
  `$m[$a->id()] = $b->id();`).
- **E3** Assignments that are not direct statements of the loop body (inside
  an `if` within the loop, inside a nested closure statement list, in the
  loop header, …).
- **E4** Left side not an array access (`$x = f($v)`, property fetch).
- **E5** Assignments outside any `for`/`foreach` body (including `while`,
  `do-while`).
- **E6** Unstable calls: `r` (itself or any sub-expression, closure bodies
  excluded) contains an assignment, an increment/decrement, a call-time
  `&` argument, an argument bound to a by-reference parameter of a resolved
  function, or a call to a built-in function that changes state or returns a
  different value on each evaluation — internal-pointer functions (`next`,
  `prev`, `reset`, `end`, `each`), `array_shift`/`array_pop`/
  `array_push`/`array_unshift`/`array_splice`, `shuffle`, `str_shuffle`,
  `array_rand`, random generators (`rand`, `mt_rand`, `random_int`,
  `random_bytes`, `lcg_value`, `uniqid`, `openssl_random_pseudo_bytes`,
  `mcrypt_create_iv`), clocks (`time`, `microtime`, `hrtime`), stream
  reads/writes (`fgets`, `fgetc`, `fgetss`, `fread`, `fgetcsv`, `fscanf`,
  `fwrite`, `fputs`, `fputcsv`, `readline`, `stream_get_line`,
  `stream_get_contents`, `socket_read`, `file_put_contents`) and
  `func_get_args`. Function names are matched on the resolved, global name,
  case-insensitively. Method calls are not inspected for side effects (their
  bodies are unknown) and stay reportable.

## Report

- Range: the full right-hand call node `r` (for a method call, from the
  start of its receiver expression to the closing `)`; for a function call,
  from the name to `)`).
- Severity: warning.
- Message: `This call is repeated in the key; store its result in a local
  variable.`

## Fix

None.

## Options

None.

## PHP versions

No gating.

## Examples

```php
<?php
foreach ($users as $user) {
    $byEmail[$user->email()] = <warning descr="This call is repeated in the key; store its result in a local variable.">$user->email()</warning>;
    $byName[strtolower($user)] = <warning descr="This call is repeated in the key; store its result in a local variable.">strtolower($user)</warning>;
    $byName[(strtolower($user))] .= <warning descr="This call is repeated in the key; store its result in a local variable.">strtolower( $user )</warning>;
    $tagged['u:' . md5($user)] = $cache[<warning descr="This call is repeated in the key; store its result in a local variable.">md5($user)</warning>];
    $groups[Role::of($user)][] = <warning descr="This call is repeated in the key; store its result in a local variable.">Role::of($user)</warning>;

    $byName[strtolower($user)] = true;
    $plain['x'] = strtolower($user);
    $other[md5($user)] = md5($guest);
    if ($user) {
        $skip[md5($user)] = md5($user);
    }
}

for ($i = 0; $i < 9; $i++) {
    $sq[abs($i)] = <warning descr="This call is repeated in the key; store its result in a local variable.">abs($i)</warning>;
}

while ($row = next($rows)) {
    $ids[key($row)] = key($row);
}

foreach ($streams as $h) {
    $lines[fgets($h)] = fgets($h);
    $pairs[next($h)] = next($h);
}
```

## Divergences

- **Name case (custos diverges):** upstream compares the called names
  case-sensitively, so `$m[f($x)] = F($x);` or
  `$m[$u->getEmail()] = $u->GetEmail();` is missed although both sides
  call the same function or method. custos compares function, method and
  class names case-insensitively, as PHP does (D5).
- **Unstable calls (custos diverges):** upstream also reports calls whose
  two evaluations differ or that change state (`$m[next($it)] = next($it);`,
  `$m[mt_rand()] = mt_rand();`). Computing them once would change the
  program, so the advice is wrong; custos skips them (E6).
- **Parenthesised assignments (custos diverges):** upstream skips an
  assignment statement wrapped in parentheses (`($m[f()] = f());`) because
  the top-level expression is not an assignment. The parentheses change
  nothing, so custos looks through them (D2) and reports the repeated call.
