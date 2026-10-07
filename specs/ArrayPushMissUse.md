---
id: ArrayPushMissUse
group: Performance
kind: semantic
needs: [types]
php: { min: "", max: "" }
---

# ArrayPushMissUse

## Summary
Appending a single element with `array_push($list, $item);` costs a function
call; the append operator `$list[] = $item;` does the same faster. Likewise,
`$list[count($list)] = $item;` computes an index that the append operator
would produce anyway (for list-shaped arrays).

## Detection

### A. Single-element `array_push` statement
- **D1** A function call (not a method/static call) that resolves to the
  global function (names compared case-insensitively, as PHP does; `\` and a
  global `use function` import are fine, but a same-named function declared in
  the current namespace, one imported from another namespace, or a qualified
  non-global name such as `Ns\array_push` does not count): `array_push`.
- **D2** It has exactly **2** arguments.
- **D3** The call is used as a statement on its own: its direct parent is an
  expression statement (`array_push($a, $b);`). Used inside an assignment,
  a condition, an argument, parentheses, `return`, etc. → no report.
- **D4** The 2nd argument is not an unpacked argument (`...$items`; whitespace
  between `...` and the expression allowed).
- **D8** The 1st argument is known to be an array: its inferred type is
  non-empty and every member is an array form (`array`, `T[]`, e.g. from a
  `$a = [];` assignment, an `array` parameter or property type, or an
  `@var T[]` doc type). Unknown, nullable, object (`ArrayAccess`), string or
  mixed types are not reported.

### B. Index computed with `count()` (option `REPORT_EXCESSIVE_COUNT_CALLS`, default on)
- **D5** An array access `C[I]` that is the **left side** of a plain
  assignment (`=`, including `= &`; compound operators such as `.=`/`+=`
  excluded). The array access must be the direct left operand (in
  `$a[count($a)][] = …` the inner access is not the left operand → no
  report).
- **D6** `I` (no parentheses around it) is a plain function call (not a
  method call) that resolves to the global `count` (case-insensitive, same
  resolution as D1) with exactly 1 argument.
- **D7** `C` is equivalent to that argument: same node kind and structurally
  identical (whitespace ignored; plain variables compare by name), e.g.
  `$this->list[count($this->list)]`.
- **Name case.** Wherever this rule compares two expressions for
  equivalence, the names PHP resolves case-insensitively — function and
  method names, class names in calls, `new`, `instanceof` and `::`
  accesses, and keywords — compare case-insensitively (`Cache::$map['k']`
  matches `cache::$map['k']`, `$o->Name()` matches `$o->name()`); variable,
  property and constant names stay case-sensitive.

## Exceptions (no report)
- **E1** `array_push` with 1 or ≥ 3 arguments, or with an unpacked 2nd
  argument.
- **E2** `array_push` whose return value is used (assigned, compared,
  passed on).
- **E6** `array_push` whose 1st argument is not known to be an array (D8).
- **E3** `count()` index not alone (`count($a) + 1`, `count($a) - 1`),
  counting a different array, `count($a, COUNT_RECURSIVE)`, wrapped in
  parentheses.
- **E4** `C[count(C)]` read (right side, argument, …) or used with a compound
  assignment.
- **E5** Option `REPORT_EXCESSIVE_COUNT_CALLS` off disables part B.

## Report
- A: range = the whole `array_push(...)` call (name through closing `)`, not
  the `;`). Severity: warning. Message: `Use '{replacement}' instead; it
  avoids a function call.` with `{replacement}` as in F1.
- B: range = **only the function name identifier** `count` of the index call
  (not its arguments). Severity: warning (upstream renders it in the
  "unused symbol" greyed style, but the severity level is the inspection
  default). Message: `The index is redundant here; use '[]' to append.`

## Fix
- **F1** (part A only) Replace the call with `{arg1}[] = {arg2}`, where
  `{arg1}`/`{arg2}` are the verbatim source texts of the 1st and 2nd
  arguments, joined exactly as `[] = ` (no space before `[]`, one space on
  each side of `=`). The trailing `;` of the statement stays.
  `array_push($queue, $job);` → `$queue[] = $job;`
  `array_push($log, [1, 2]);` → `$log[] = [1, 2];`
- Part B has no fix.

## Options
| Option | Type | Default | Effect |
|---|---|---|---|
| REPORT_EXCESSIVE_COUNT_CALLS | bool | true | Enables part B (`C[count(C)] = …`). |

## PHP versions
No gating. (Unpacking `...` requires 5.6 at runtime; it is merely recognised
as an exclusion.)

## Examples

```php
<?php
$queue = [];
$log = [];
<warning descr="Use '$queue[] = $job' instead; it avoids a function call.">array_push($queue, $job)</warning>;
<warning descr="Use '$log[] = [1, 2]' instead; it avoids a function call.">\array_push($log, [1, 2])</warning>;
array_push($unknown, $job);

array_push($queue, $a, $b);
array_push($queue, ... $more);
$size = array_push($queue, $job);
while (array_push($queue, $job) < 10) {}

$queue[<warning descr="The index is redundant here; use '[]' to append.">count</warning>($queue)] = $job;
$this->items[<warning descr="The index is redundant here; use '[]' to append.">count</warning>($this->items)] = $job;

$queue[count($queue) - 1] = $job;
$queue[count($other)] = $job;
$last = $queue[count($queue)];
$queue[count($queue)] .= 'x';
```

```php
<?php
$queue = [];
$log = [];
$queue[] = $job;
$log[] = [1, 2];
array_push($unknown, $job);

array_push($queue, $a, $b);
array_push($queue, ... $more);
$size = array_push($queue, $job);
while (array_push($queue, $job) < 10) {}

$queue[count($queue)] = $job;
$this->items[count($this->items)] = $job;

$queue[count($queue) - 1] = $job;
$queue[count($other)] = $job;
$last = $queue[count($queue)];
$queue[count($queue)] .= 'x';
```

## Divergences
- custos diverges from upstream on the target type (D8, E6). Upstream
  reports every single-element `array_push` statement, but `$a[] = $b` is
  not the same operation when `$a` is not an array: on an `ArrayAccess`
  object it calls `offsetSet(null, $b)` instead of failing, on `null` it
  silently creates an array where `array_push` throws, and on a string it
  is an error of a different kind. custos therefore reports (and fixes)
  only when the first argument is known to be an array. The EA
  conformance case still passes.
- Part B with a qualified call (`\count($a)`): whether upstream's range
  includes the leading `\` is unverified; recommendation: highlight the
  name identifier only (without the qualifier).
- Named arguments (`array_push(array: $a, values: $b)`) are not
  distinguished upstream and would yield `array: $a[] = values: $b`;
  recommendation: skip calls with named arguments. No fixture covers it.
- **Function-name matching — custos diverges from upstream.** Upstream matches
  `array_push` and `count` by the name as written (case-sensitive, any
  namespace qualifier, no resolution), so a differently cased call such as
  `Array_Push($a, $b)` is missed while a namespaced or imported user function
  of the same name is reported (and rewritten) as if it were the builtin.
  custos matches case-insensitively and only calls that reach the global
  function.
- **Name case (custos diverges).** Upstream compares the expressions
  textually, so operands that differ only in the case of a function, method
  or class name (`Stats::$n` vs `stats::$n`), which PHP treats as the same,
  are not recognised as equivalent. custos folds the case of those names
  (Detection, "Name case").
