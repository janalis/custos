---
id: UntrustedInclusion
group: Security
kind: semantic
needs: [names, index]
php: { min: "", max: "" }
---

# UntrustedInclusion

## Summary
Including a file by a relative path makes PHP search `include_path` (and the
current working directory), so a different file than intended can be loaded.
Anchor paths with `__DIR__` (or rely on autoloading).

## Detection
- **D1** Any inclusion expression: `include`, `include_once`, `require`,
  `require_once`, with or without parentheses around the argument, in any
  context (statement, assignment, condition…).
- **D2** The argument resolves to a single string literal: the argument
  itself if it is a string literal; otherwise run *value discovery* (as
  defined in the `CallableMethodValidity` spec — parentheses are stripped
  first) and keep the discovered string literals; there must be exactly
  one.
- **D3** The literal's raw content (between the quotes, escapes not
  decoded) is non-empty and does **not** start with `/`, with a backslash `\` (a UNC
  path `\\host\share\x.php` or a Windows drive-root path), with a single
  ASCII letter followed by `:` (case-insensitive, e.g. `C:`, `d:`), or with
  a stream-wrapper scheme (E4) → report. Explicitly relative paths (`./x.php`,
  `../x.php`) are still reported: they depend on the working directory.

## Exceptions (no report)
- **E1** Absolute paths: `'/etc/app.php'`, `'C:\app\x.php'`, `'c://x.php'`.
- **E2** Arguments that do not resolve to exactly one string literal:
  concatenations (`__DIR__ . '/x.php'`), constants, function calls,
  variables at top level, variables with several possible literal values.
- **E3** Empty string `''`.
- **E4** Stream-wrapper paths: the content starts with a scheme of two or
  more characters (an ASCII letter, then letters, digits, `+`, `-` or `.`)
  followed by `://` (`'phar://app.phar/boot.php'`, `'file:///srv/x.php'`).
- **E5** UNC and drive-root paths starting with a backslash
  (`'\\\\host\\share\\x.php'`).

## Report
- Range: the whole inclusion expression, from the keyword to the end of the
  argument (closing `)` included when parenthesised; the terminating `;`
  excluded).
- Severity: error (the rule is disabled by default).
- Message: `Relative include depends on include_path; anchor it with __DIR__.`

## Fix
None.

## Options
None.

## PHP versions
No gating.

## Examples

```php
<?php
<error descr="Relative include depends on include_path; anchor it with __DIR__.">require_once 'lib/boot.php'</error>;
<error descr="Relative include depends on include_path; anchor it with __DIR__.">include ("views/header.phtml")</error>;
$cfg = <error descr="Relative include depends on include_path; anchor it with __DIR__.">require('settings.php')</error>;

function plugin()
{
    $entry = 'plugins/main.php';
    <error descr="Relative include depends on include_path; anchor it with __DIR__.">include_once $entry</error>;
}

require __DIR__ . '/lib/boot.php';
include '/srv/app/shared.php';
require 'D:/apps/shared.php';
include $dynamic;
include '';
include 'phar://app.phar/boot.php';
include '\\\\fileserver\\shared\\boot.php';
```

## Divergences
- **E4/E5 — custos diverges from upstream.** Upstream reports stream-wrapper
  paths (`'phar://app.phar/x.php'`, `'file:///x.php'`) and UNC paths
  (`'\\\\host\\x.php'`) as relative includes. Neither is resolved through
  `include_path`, and anchoring them with `__DIR__` would break them, so
  custos does not report them. `./x.php` / `../x.php` stay reported (they
  are relative to the working directory, same as upstream).
- Double-quoted strings with interpolation (`"$base/x.php"`) are string
  literals whose raw content does not start with `/`, so upstream reports
  them although the real path is unknown. Recommendation: do not report a
  literal whose content starts with an interpolation; not covered by
  fixtures.
