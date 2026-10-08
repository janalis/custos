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
  between the backticks:
  1. every escaped backtick `` \` `` becomes a plain backtick `` ` ``;
  2. every double quote `"` becomes `\"`.
  The rest of the content (spaces, variables, `{$...}` interpolations,
  other backslash sequences) is copied verbatim — see Divergences.
  Examples:
  - `` `uptime` `` → `shell_exec("uptime")`
  - `` `grep "needle" log.txt` `` → `shell_exec("grep \"needle\" log.txt")`
  - `` `echo \`date\`` `` → ``shell_exec("echo `date`")``
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
  rules: `\n` is a newline, `$var` interpolates). Recommendation: copy all
  other characters verbatim (semantics-preserving), escape only a `"` that
  is not already escaped (preceded by an even number of backslashes,
  including zero). Not covered by fixtures beyond `"`.
- The identifier-boundary space in F1 is our addition (upstream's
  replacement is a PSI node so the question does not arise); the fixture
  case has `return` followed by a space.
- **Builtin spelling (custos diverges).** Upstream inserts a bare `shell_exec(`,
  which a function of that name declared in (or imported into) the current
  namespace captures, so the fix would call user code. custos writes
  `\shell_exec(` in that case.
