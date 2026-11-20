---
id: Formatting
group: Code style
kind: syntax
needs: []
php: { min: "", max: "" }
---

# Formatting

## Summary

Reports every place where the file's layout differs from what `custos format`
would write: indentation, spaces, line breaks, braces, blank lines and the
other settings of the project's format configuration. It brings formatting
checks into the analysis, the editor and the baseline, with each change
offered as a quick-fix.

## Detection

- D1. Lay the file out as `custos format` would, with the format settings that
  apply to it: the `format` section of the configuration, `.editorconfig`
  files and the configuration of other tools, as for the `format` command
  without command-line options. Report each change the formatter makes.
- D2. When the format settings ask for long lines to be reported, report each
  line that stays wider than the line length once formatted.

## Exceptions (no report)

- E1. Files the `format.exclude` globs leave out, and every file when the
  format settings are invalid.
- E2. Files the formatter leaves as written: files with syntax errors, files
  carrying a format-ignore comment, and files whose formatting the formatter
  cannot prove safe.
- E3. Code inside format-ignore regions.

## Report

- Range: the source text the change replaces; an empty range where it only
  inserts text. For a long line, an empty range at its first token.
- Severity: info.
- Message: what the change does, such as `Fix the indentation` or
  `Move the function opening brace`; the same text `custos format --check`
  prints.

## Fix

- F1. The first quick-fix applies that one change.
- F2. The second quick-fix, `Format the file`, applies every change of the
  file at once, giving the output of `custos format`.

A long line (D2) has no quick-fix: formatting cannot shorten it.

## Options

None.

## PHP versions

No gating. The formatter parses the file at the target PHP version.

## Examples

```php
<?php

function total($a,<info descr="Fix the spaces between tokens"></info>$b)<info descr="Move the function opening brace"></info>{<info descr="Fix the indentation">
  </info>return $a + $b;
}
```

```php
<?php

function total($a, $b)
{
    return $a + $b;
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA
conformance requirement.
