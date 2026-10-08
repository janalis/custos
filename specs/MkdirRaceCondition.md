---
id: MkdirRaceCondition
group: Probable bugs
kind: semantic
needs: [names, stubs]
php: { min: "", max: "" }
---

# MkdirRaceCondition

## Summary

`mkdir()` fails when the directory already exists — including when another
process created it a moment after our own `is_dir()` check. A failed
`mkdir()` must therefore be followed by an `is_dir()` re-check before being
treated as an error, and its result must not be ignored.

## Detection

- **D1** A plain function call whose name is `mkdir`, compared case-insensitively as PHP
  does (`MkDir(...)` qualifies; custos diverges),
  with 1 to 3 arguments (positional or named), resolving to the global
  built-in `mkdir` (no user function `mkdir` declared in the current
  namespace; an unqualified call in a namespace falls back to the global
  one). Method calls and `new` are never considered.
- **D2** Not a test context: the file path does not end with `Test.php`,
  `Spec.php` or `.phpt` and does not contain `/Fixtures/`, and the enclosing
  class (if any) FQN does not end with `Test` nor contain `\Tests\` or
  `\Test\`.
- **D3** Locate the *target* by walking upwards from the call, starting with
  `inverted = false`. At each step look at the parent `P` of the current node
  `N`:
  - `P` is an `if` whose condition is `N`, or an assignment expression, or an
    expression statement → the target is `N`; stop.
  - `P` is a parenthesised expression → continue with `N = P`.
  - `P` is unary `!` → flip `inverted`, continue with `N = P`.
  - `P` is the error-suppression `@` → continue with `N = P`.
  - `P` is any other unary operator (cast, `-`, `~`, …) → no target; stop.
  - `P` is a logical binary (`&&`, `||`, `and`, `or`) → the target is `N`;
    stop.
  - `P` is any other binary operator → if it is `===` or `!==` and the *other*
    operand is a boolean literal (`true`/`false`, any case): flip `inverted`
    when that literal is `false`, and flip again when the operator is `!==`.
    (Other operators such as `==` do not flip.) Continue with `N = P`.
  - Anything else (`elseif`/`while` condition, argument, `return`, `echo`,
    ternary, array element, …) → no target; stop, no report.
- Let `T` be the target and `C` its parent.
- **D4** `C` is an expression statement → report the statement (result
  ignored). Fix F1.
- **D5** `C` is an `if` (T is its whole condition) → report `T`; the
  suggested form is the *and*-form when `inverted` is true, else the
  *or*-form. Fix F2.
- **D6** `C` is a logical binary expression `B`:
  - if `B`'s right operand is `exit`/`die` (`mkdir(...) or die(...)`) → no
    report;
  - if `T` is `B`'s right operand and `B`'s parent is itself a binary
    expression, use that parent as `B'` for the next check, else `B' = B`;
  - look at `B'`'s right operand and every function call nested in it; if
    any is a plain function call named `is_dir` (case-insensitive) whose
    first argument re-checks the same directory → no report. The first
    argument matches when its text (outer parentheses and whitespace
    ignored) equals the text of `mkdir()`'s first argument, or, when that
    argument is an assignment `$v = expr`, equals `$v` or `expr`. An
    `is_dir()` on another path does not count;
  - otherwise report `T`; the suggested form is the *and*-form when `B'`'s
    operator is `&&`/`and`, the *or*-form when it is `||`/`or`. Fix F3.
- **D7** Any other `C` (e.g. assignment `$ok = mkdir(...)`) → no report.

## Exceptions (no report)

- **E1** Result stored: `$ok = mkdir($d);` (also inside parentheses or after
  `!`/`@`).
- **E2** `mkdir(...) or die(...)` / `or exit`.
- **E3** An `is_dir()` re-check of the same directory in the right operand
  of the (outer) logical expression: `!is_dir($d) && !mkdir($d) && !is_dir($d)`,
  `mkdir($d) || is_dir($d)`, `!mkdir($d = $cfg->path()) && !is_dir($d)`.
- **E4** Call with 0 or more than 3 arguments; non-global `mkdir`; test
  context.
- **E5** Calls whose result flows anywhere other than D4–D6 (returned,
  passed as an argument, `elseif`/`while` conditions, ternaries, casts).

## Report

- Range:
  - D4: the whole expression statement, including a leading `@`/`!` and the
    terminating `;`.
  - D5/D6: the target node `T` exactly — including any parentheses, `!`,
    `@` and comparison wrapped around the call (`(mkdir($d) !== false)` when
    written as `if ((mkdir($d) !== false))`; `!mkdir($d)` in
    `!is_dir($d) && !mkdir($d)`). Parentheses belonging to the `if` itself are
    not included.
- Severity: error.
- Messages (`{args}` = the call's argument list text, between its
  parentheses, verbatim):
  - D4: `mkdir() outcome is ignored; use 'if (!mkdir({args}) && !is_dir(...)) { ... }'.`
  - and-form: `Re-check with is_dir() after a failed mkdir: '!mkdir({args}) && !is_dir(...)'.`
  - or-form: `Re-check with is_dir() after a failed mkdir: 'mkdir({args}) || is_dir(...)'.`

## Fix

Notation: `ARGS` = argument list text verbatim (e.g. `$path, 0755, true` or
`$path, recursive: true`); `DIR` = text of the first argument. A *temp-var*
form is used when the first argument is neither a plain variable nor a string
literal (any kind of quoted string): then `ARGS` is used as
`$concurrentDirectory = ARGS` inside the `mkdir(...)` call (which assigns only
the first argument, since the comma ends the assignment) and
`$concurrentDirectory` replaces `DIR` afterwards.

Builtin spelling: in every form below, `mkdir`, `is_dir` and `sprintf` are
written with a leading `\` when an unqualified call at the reported
position would not reach the global function (a `use function` import under
that name, or a same-named function declared in the current namespace);
`mkdir` also keeps the leading `\` of an original `\mkdir(...)`. E.g. in a
namespace declaring `function is_dir()`, F1 gives
`if (!mkdir(ARGS) && !\is_dir(DIR)) { … }`.

- **F1** (D4) Replace the whole statement with an `if` that throws. Let
  `FMT` be the single-quoted PHP literal whose contents are
  `Directory "%s" was not created` (with the double quotes and the `%s`
  literally), and `THROW(X)` = `throw new \RuntimeException(sprintf(FMT, X));`
  (one space after the comma inside `sprintf`):
  - normal: `if (!mkdir(ARGS) && !is_dir(DIR)) { THROW(DIR) }`
  - temp-var: `if (!mkdir($concurrentDirectory = ARGS) && !is_dir($concurrentDirectory)) { THROW($concurrentDirectory) }`
  Any `!` in front of the original call is dropped; an `@` is kept in front
  of the new `mkdir` (`!@mkdir(ARGS) && !is_dir(DIR)`; also in F2/F3, see
  Divergences). Keep a space after
  `{` and before `}` (output is compared whitespace-collapsed).
- **F2** (D5) Replace the whole `if` condition `T` (dropping its extra
  parentheses, comparisons, `!` and `@`) with:
  - normal, inverted: `!mkdir(ARGS) && !is_dir(DIR)`
  - normal, not inverted: `mkdir(ARGS) || is_dir(DIR)`
  - temp-var, inverted: `!mkdir($concurrentDirectory = ARGS) && !is_dir($concurrentDirectory)`
  - temp-var, not inverted: `mkdir($concurrentDirectory = ARGS) || is_dir($concurrentDirectory)`
    (the `is_dir` re-check is not negated, like the normal or-form; see
    Divergences).
- **F3** (D6) Only when `T` is the **right** operand of its parent binary `B`
  (the original parent, not `B'`). Replace `B` with `L OP M` where `L` is the
  text of `B`'s left operand, verbatim:
  - `B` is `&&`/`and`: `L && !mkdir(ARGS) && !is_dir(DIR)` (temp-var:
    `L && !mkdir($concurrentDirectory = ARGS) && !is_dir($concurrentDirectory)`).
  - `B` is `||`/`or`: `L || mkdir(ARGS) || is_dir(DIR)` (temp-var:
    `L || mkdir($concurrentDirectory = ARGS) || is_dir($concurrentDirectory)`).
  The operator is always written `&&`/`||`, regardless of `inverted` or of
  `!`/`@` around the original call. If `B` sits inside a larger expression
  the replacement is inserted as-is (no added parentheses).
  When `T` is the left operand, no fix is offered (see Divergences).

## Options

None.

## PHP versions

None. Named arguments (PHP 8.0) appear in the upstream fixture, which runs at
the test default level; parse them regardless of level.

## Examples

```php
<?php
function prepare($root, $cfg) {
    <error descr="mkdir() outcome is ignored; use 'if (!mkdir($root, 0750) && !is_dir(...)) { ... }'.">@mkdir($root, 0750);</error>
    $made = mkdir("$root/cache");
    mkdir($root . '/tmp') or exit(1);

    if (<error descr="Re-check with is_dir() after a failed mkdir: '!mkdir($root) && !is_dir(...)'.">(false === mkdir($root))</error>) {}
    if (<error descr="Re-check with is_dir() after a failed mkdir: 'mkdir(&quot;$root/log&quot;, 0700, true) || is_dir(...)'.">(mkdir("$root/log", 0700, true))</error>) {}
    if (<error descr="Re-check with is_dir() after a failed mkdir: 'mkdir(dirname($root)) || is_dir(...)'.">(mkdir(dirname($root)))</error>) {}
    if (file_exists($root) || <error descr="Re-check with is_dir() after a failed mkdir: 'mkdir($root, permissions: 0700) || is_dir(...)'.">mkdir($root, permissions: 0700)</error>) {}
    if ($cfg->ok && <error descr="Re-check with is_dir() after a failed mkdir: '!mkdir(strtolower($root)) && !is_dir(...)'.">!mkdir(strtolower($root))</error>) {}

    if (!is_dir($root) && !mkdir($root) && !is_dir($root)) {}
    return mkdir($root);
}
```

```php
<?php
function prepare($root, $cfg) {
    if (!mkdir($root, 0750) && !is_dir($root)) { throw new \RuntimeException(sprintf('Directory "%s" was not created', $root)); }
    $made = mkdir("$root/cache");
    mkdir($root . '/tmp') or exit(1);

    if (!mkdir($root) && !is_dir($root)) {}
    if (mkdir("$root/log", 0700, true) || is_dir("$root/log")) {}
    if (mkdir($concurrentDirectory = dirname($root)) || is_dir($concurrentDirectory)) {}
    if (file_exists($root) || mkdir($root, permissions: 0700) || is_dir($root)) {}
    if ($cfg->ok && !mkdir($concurrentDirectory = strtolower($root)) && !is_dir($concurrentDirectory)) {}

    if (!is_dir($root) && !mkdir($root) && !is_dir($root)) {}
    return mkdir($root);
}
```

A statement `mkdir($cfg->dir());` would become the temp-var form of F1
(`$concurrentDirectory = $cfg->dir()` inside the call, `$concurrentDirectory`
in `is_dir` and in `THROW`).

(`(false === mkdir($root))`: the boolean is the other operand, so it flips to
inverted.)

## Divergences

- **Name case (custos diverges).** Upstream matches the written name `mkdir`
  case-sensitively, so `MKDIR($d)` used as a condition escapes the check
  although PHP calls the same builtin. custos compares it
  case-insensitively (D1).
- F2 temp-var, not inverted: custos diverges from upstream. Upstream emits
  `mkdir($concurrentDirectory = ARGS) || !is_dir($concurrentDirectory)`,
  which inverts the re-check: the condition becomes true when `mkdir()`
  succeeds or when the directory does **not** exist, i.e. exactly when the
  creation failed. custos emits `|| is_dir($concurrentDirectory)`, the same
  shape as the normal or-form, so the condition still means "the directory
  is there". The upstream fixture expecting the negated form is a listed
  conformance divergence.
- Upstream also offers F3 when the call is the **left** operand of the logical
  expression (`mkdir($d) && log()`); its fix then rebuilds the expression from
  the left operand only, duplicating `mkdir` and dropping the right operand.
  Recommendation: report but offer no fix in that case (no fixture covers it).
- `is_dir()` re-check on another path (custos diverges from upstream).
  Upstream accepts any `is_dir` call in the right operand, whatever its
  argument, so `!mkdir($cache) && !is_dir($logs)` passes although the
  created directory is never re-checked. custos requires the re-check to
  name the same directory as `mkdir()` (D6), and matches `is_dir`
  case-insensitively like PHP does.
- **Builtin spelling (custos diverges).** Upstream writes bare `mkdir(`,
  `is_dir(` and `sprintf(` in its fixes, so a function of that name declared
  in (or imported into) the namespace captures the re-check, and an
  original `\mkdir(...)` in a namespace declaring its own `mkdir` is
  rewritten to call the user function. custos qualifies those names with
  `\` when needed (Fix, "Builtin spelling").
- **Silence operator kept (custos diverges).** Upstream drops an `@` in
  front of the original call. Without it, `mkdir()` emits a "File exists"
  warning precisely in the race the re-check is meant to absorb (and
  frameworks that turn warnings into exceptions then throw). custos writes
  `@mkdir(ARGS)` in every fix form when the original call was silenced.
- **First-class callable (custos fix).** `mkdir(...)` builds a closure and
  creates nothing; it is not reported (D1). An earlier custos version
  reported `mkdir(...);` as an ignored outcome and offered a fix producing
  `mkdir($concurrentDirectory = ...)`, which does not parse.
- **Re-check in the failure branch (custos diverges).**
  `if (!mkdir($d)) { clearstatcache(); if (is_dir($d)) { return; } … }`
  already re-checks the same directory inside the `if` body (Composer);
  D5 no longer reports the and-form when that body contains an `is_dir()`
  call on the directory (same matching as D6).
- **Re-check in a later statement (custos diverges).** A statement-form
  `mkdir($d);` (D4) followed, in the same function, by a statement calling
  `is_dir($d)` (`if (!is_dir($p)) { mkdir($p, 0777, true); } return
  is_dir($p) && is_writable($p);`, CakePHP) already handles the failure;
  the D4 fix would turn a `false` result into an exception. Not reported
  (same matching as D6; following statements of the enclosing statement
  lists up to the function body). Besides `is_dir()`, a later
  `file_exists()`, `realpath()`, `is_writable()` or `is_writeable()` of the
  same directory counts as the check (`@mkdir($p); $real = realpath($p);`
  then a test of `$real`, Kimai's doctor page). The scan stops at a
  statement that assigns the directory expression (`$dir .= '/sub';`): a
  later check then names another directory.
- **Polarity in logical operands (custos diverges, D6/F3).** The re-check
  form follows the call's polarity, not the operator: `A || !mkdir($d)`
  becomes `A || (!mkdir($d) && !is_dir($d))` and `A && mkdir($d)` becomes
  `A && (mkdir($d) || is_dir($d))` (parenthesised when the form's operator
  differs); upstream emitted the operator's form and dropped the `!`,
  running the error branch on success (Joomla image thumbnails). The
  message follows the polarity too, and the operator is kept as written
  (`or` binds looser than an assignment).
- **Directory locks (custos diverges).** `if (!@mkdir($lock)) { return; }
  … rmdir($lock);` uses mkdir() as an atomic lock: its failure means
  "someone else holds the lock", and the suggested `&& !is_dir($lock)`
  re-check would make every process take it (FreshRSS's migrator). A
  mkdir() whose outcome is tested (not a bare statement) and whose
  function (or top-level code) also calls `rmdir()` on the same directory
  expression is not reported; an ignored `mkdir($scratch); …
  rmdir($scratch);` temporary directory still is.
- **Retry loops (custos).** A `mkdir()` in the condition of a `while` /
  `do … while` loop whose condition also tests `is_dir()` of the same
  directory (`while (!is_dir($d) && !@mkdir($d, 0755, true)) { usleep(…); }`)
  re-checks on the next iteration: not reported (October CMS's
  `CodeParser`).
- **Unbraced bodies (custos).** When the ignored `mkdir()` statement is the
  unbraced body of an `if`/`else`/loop, the D4 replacement is wrapped in
  braces: a bare `if (…) { throw … }` would capture the outer `else`.
