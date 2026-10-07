---
id: ShortEchoTagCanBeUsed
group: Code style
kind: syntax
needs: []
php: { min: "", max: "" }
---

# ShortEchoTagCanBeUsed

## Summary
In templates, a PHP block that only outputs something — `<?php echo $x ?>` —
reads better as the short echo tag `<?= $x ?>`.

## Detection
Token-level view: a PHP block delimited by an opening tag and a closing tag
that contains exactly one statement.

- D1: an `echo` statement (one or more comma-separated arguments; with or
  without a trailing `;`) — or an expression statement consisting solely of a
  `print` expression (with or without a trailing `;`).
- D2: immediately before the statement there is a regular opening tag
  (`<?php`, any letter case, or the short `<?` open tag; NOT `<?=`), with
  nothing in between except whitespace (no comments, no other statements).
- D3: immediately after the statement (after its `;` if present) there is a
  closing tag `?>`, with nothing in between except whitespace.
- Report when D1–D3 hold.

## Exceptions (no report)
- E1: the block contains more than one statement (`<?php echo $a; echo $b; ?>`
  — neither is reported).
- E2: already short echo tags (`<?= … ?>`).
- E3: a comment between the opening tag and the statement, or between the
  statement and `?>`.
- E4: no closing tag after the statement (e.g. end of file).
- E5: `print` used inside a larger expression (`<?php $r = print $x ?>`).

## Report
- Range: the `echo` or `print` keyword token only.
- Severity: info (weak warning).
- Message: "Use the short echo tag '<?= … ?>' here."

## Fix
- F1: replace the opening tag token with `<?=`; replace the statement (from the
  keyword through its trailing `;` if any) with its argument texts joined by
  `, ` (for `print`, its single argument's text). The trailing `;` is dropped.
  Whitespace between the opening tag and the statement, and between the
  statement and `?>`, is preserved. Argument texts are copied verbatim
  (parentheses included, e.g. `print($v)` → `($v)`).
  `<?php echo $a, $b; ?>` → `<?= $a, $b ?>`.

## Options
None.

## PHP versions
No gating (the rule is disabled by default). `<?=` is always available from
PHP 5.4.

## Examples

```php
<ul>
<li><?php <weak_warning descr="Use the short echo tag '<?= … ?>' here.">echo</weak_warning> $item->name; ?></li>
<li><?php <weak_warning descr="Use the short echo tag '<?= … ?>' here.">echo</weak_warning> strtoupper($code) ?></li>
<li><?PHP <weak_warning descr="Use the short echo tag '<?= … ?>' here.">echo</weak_warning> $first, ' ', $last; ?></li>
<li><?php <weak_warning descr="Use the short echo tag '<?= … ?>' here.">print</weak_warning> $total; ?></li>
<li><?php <weak_warning descr="Use the short echo tag '<?= … ?>' here.">print</weak_warning>($note) ?></li>
<li><?= $already ?></li>
<li><?php echo $a; echo $b; ?></li>
<li><?php /* hint */ echo $c ?></li>
</ul>
```

```php
<ul>
<li><?= $item->name ?></li>
<li><?= strtoupper($code) ?></li>
<li><?= $first, ' ', $last ?></li>
<li><?= $total ?></li>
<li><?= ($note) ?></li>
<li><?= $already ?></li>
<li><?php echo $a; echo $b; ?></li>
<li><?php /* hint */ echo $c ?></li>
</ul>
```

## Divergences
None known.
