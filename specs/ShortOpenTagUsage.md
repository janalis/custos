---
id: ShortOpenTagUsage
group: Code style
kind: syntax
needs: []
php: { min: "", max: "" }
---

# ShortOpenTagUsage

## Summary
The short open tag `<?` depends on the `short_open_tag` ini setting; code using
it breaks (prints source code) where that setting is off. Use `<?php`.

## Detection
Token-level: requires the lexer to recognise `<?` as an opening tag (as PHP
does with `short_open_tag=On`).

- D1: an opening-tag token whose text is exactly `<?` (not `<?php` in any
  case, not `<?=`).
- D2: the token is immediately followed by whitespace (space, tab, newline), or
  it is the last token of the file.

## Exceptions (no report)
- E1: `<?php` and `<?=` tags.
- E2: `<?` immediately followed by a non-whitespace character (e.g. `<?xml`,
  `<?echo`): not reported (also avoids turning `<?echo` into `<?phpecho`).

## Report
- Range: the two characters `<?`.
- Severity: info (weak warning).
- Message: "Replace the short open tag '<?' with '<?php'."

## Fix
- F1: replace the `<?` token with `<?php`; the following whitespace and code
  are unchanged.

## Options
None.

## PHP versions
No gating.

## Examples

```php
<weak_warning descr="Replace the short open tag '&lt;?' with '&lt;?php'."><?</weak_warning> $greeting = 'hi'; ?>
<p>
<weak_warning descr="Replace the short open tag '&lt;?' with '&lt;?php'."><?</weak_warning> foreach ($users as $u) : ?>
  <b><?= $u ?></b>
<weak_warning descr="Replace the short open tag '&lt;?' with '&lt;?php'."><?</weak_warning> endforeach ?>
<weak_warning descr="Replace the short open tag '&lt;?' with '&lt;?php'."><?</weak_warning>
  while (next_row()) { ?><i>row</i><?php } ?>
</p>
<?php $done = true; ?>
```

```php
<?php $greeting = 'hi'; ?>
<p>
<?php foreach ($users as $u) : ?>
  <b><?= $u ?></b>
<?php endforeach ?>
<?php
  while (next_row()) { ?><i>row</i><?php } ?>
</p>
<?php $done = true; ?>
```

## Divergences
None known.
