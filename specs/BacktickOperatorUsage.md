---
id: BacktickOperatorUsage
group: Security
kind: syntax
needs: []
php: { min: "", max: "" }
---

# BacktickOperatorUsage

## Summary

The backtick operator runs a shell command just like `shell_exec()`, but it
is easy to overlook when reading code and security scanners often miss it.
Calling `shell_exec()` explicitly makes command execution visible.

## Detection

- **D1** Every shell-command expression (backtick literal `` `...` ``),
  wherever it appears (statement, argument, `return`, assignment, inside
  interpolation contexts, …).
- **D2** Its full source text, backticks included, is longer than 2
  characters, i.e. the content between the backticks is not empty.

## Exceptions (no report)

- **E1** The empty command ` `` ` (two backticks, nothing between).
- **E3** Incomplete commands recovered from malformed input, including a final
  escaped backtick with no closing delimiter, are not reported or rewritten.
- **E2** Backticks inside ordinary string literals or comments are not shell
  commands and are never considered.

## Report

- Range: the whole backtick expression, from the opening backtick to the
  closing backtick inclusive.
- Severity: warning.
- Message: `Run the command through shell_exec() instead of backticks.`

## Fix

- **F1** Replace the whole backtick expression with
  `shell_exec("` + `C` + `")`, where `C` is computed from the raw text
  between the backticks. Only the literal text is rewritten, pair by pair
  (a backslash and the character after it form one pair):
  1. an escaped backtick `` \` `` becomes a plain backtick `` ` ``;
  2. a bare double quote `"` becomes `\"`;
  3. the pair `\"` becomes `\\\"`: backticks do not treat `\"` as an
     escape, so the command contains both the backslash and the quote, and
     both must be escaped to survive in a double-quoted string;
  4. every other pair or character (spaces, `\\`, `\$`, `\n`, …) is copied
     verbatim.
  Interpolations (`$var`, `$a[0]`, `$o->p`, `{$...}`, `${...}`), including
  their braces, are PHP code rather than string text and are copied
  untouched: a `"` inside `{$a["k"]}` is never escaped.
  Examples:
  - `` `uptime` `` → `shell_exec("uptime")`
  - `` `grep "needle" log.txt` `` → `shell_exec("grep \"needle\" log.txt")`
  - `` `echo \`date\`` `` → ``shell_exec("echo `date`")``
  - `` `echo {$a["k"]}` `` → `shell_exec("echo {$a["k"]}")`
  - `` `echo \"hi\"` `` → `shell_exec("echo \\\"hi\\\"")`
- The fix is offered for every report. Only the expression range is
  replaced; surrounding code is untouched. PHP accepts a keyword directly
  followed by a backtick (`` return`ls`; ``); to keep such code valid,
  insert a single space before `shell_exec` when the character immediately
  preceding the expression is an identifier character (letter, digit, `_`):
  `` return`ls`; `` → `return shell_exec("ls");`.
  Builtin spelling: `shell_exec` is written `\shell_exec` when an unqualified call at
  that position would not reach the global function (a `use function`
  import under that name, or a same-named function declared in the current
  namespace).

## Options

None.

## PHP versions

No gating.

## Examples

```php
<?php
$load  = <warning descr="Run the command through shell_exec() instead of backticks.">`cat /proc/loadavg`</warning>;
$found = <warning descr="Run the command through shell_exec() instead of backticks.">`grep -c "ERROR" app.log`</warning>;
function stamp() {
    return <warning descr="Run the command through shell_exec() instead of backticks.">`date +%s`</warning>;
}
$nothing = ``;
$text = 'uses `backticks` in a plain string';
```

```php
<?php
$load  = shell_exec("cat /proc/loadavg");
$found = shell_exec("grep -c \"ERROR\" app.log");
function stamp() {
    return shell_exec("date +%s");
}
$nothing = ``;
$text = 'uses `backticks` in a plain string';
```

## Divergences

- Upstream escapes the content with the IDE's generic "escape for a
  double-quoted string" helper. Only the behaviour for `"` is confirmed by
  fixtures. That helper may additionally double backslashes, escape `$`
  and turn raw newlines/tabs into `\n`/`\t`; doing so would change the
  meaning of backtick content (which already follows double-quoted string
  rules: `\n` is a newline, `$var` interpolates). custos copies all other
  characters verbatim and re-escapes only what differs between the two
  syntaxes: a bare `"`, the pair `\"` (a literal backslash plus quote in
  backticks, see F1.3) and `` \` ``. Quotes inside an interpolated
  expression are code and stay as written; escaping them would produce
  invalid PHP. Not covered by upstream fixtures beyond the bare `"`.
- The identifier-boundary space in F1 is our addition (upstream's
  replacement is a PSI node so the question does not arise); the fixture
  case has `return` followed by a space.
- **Builtin spelling (custos diverges).** Upstream inserts a bare `shell_exec(`,
  which a function of that name declared in (or imported into) the current
  namespace captures, so the fix would call user code. custos writes
  `\shell_exec(` in that case.
