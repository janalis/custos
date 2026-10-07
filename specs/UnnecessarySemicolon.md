---
id: UnnecessarySemicolon
group: Code style
kind: syntax
needs: []
php: { min: "", max: "" }
---

# UnnecessarySemicolon

## Summary
Stray semicolons — empty statements and the terminator right before `?>` in a
short echo tag — add nothing and are usually typos. Remove them.

## Detection
Skipped entirely when the file name ends with `.blade.php`.

### Empty statements
- **D1** An empty statement: a statement consisting solely of `;` (no
  expression, no keyword). Statements such as `return;`, `break;`,
  `continue;`, `declare(...);` are not empty statements.
- **D2** Its direct parent is **not** one of:
  - an `if`, `elseif`, `else if` or `else` clause (i.e. the `;` is the body of
    that clause: `if ($x) ;`, `else ;`);
  - a loop: `for`, `foreach`, `while`, `do … while` (as its body: `while (f()) ;`,
    `do ; while (...)`);
  - a `declare(...)` construct (`declare(ticks=1);`).
  Anywhere else — top level, inside a `{ }` block (function body, block of an
  `if`, loop, `switch` case list, namespace block), directly after another
  statement (`foo();;`), after a closing brace (`class A {};`) — it is reported.

### Short echo tags
- **D3** A short echo tag statement `<?= expr ;` (not an `echo` keyword
  statement) whose last token is `;`, and the next non-whitespace thing after
  it is not a PHP statement/expression (typically the closing `?>`, or a
  comment followed by `?>`). Report that `;`.
  If another statement follows in the same tag (`<?= $a; $b = 1; ?>`), nothing
  is reported.
- Regular statements before `?>` (`<?php foo(); ?>`) are never reported.

## Exceptions (no report)
- **E1** Bodies of `if`/`elseif`/`else`/loops/`declare` consisting of `;`.
- **E2** `<?= ...; ?>` followed by more statements in the same tag.
- **E3** `.blade.php` files.

## Report
- Range: the single `;` character (the whole empty statement for D1; the
  terminator token for D3).
- Severity: info.
- Message: `Stray semicolon; remove it.`

## Fix
- **F1** Delete the `;`. If the token immediately preceding it is whitespace
  (spaces, tabs, newlines — one contiguous run), delete that whitespace too.
  If something else precedes it (another `;`, `}`, a comment), only the `;` is
  removed.
  - `run();;` → `run();`
  - a line containing only `    ;` after `run();` → the newline+indent and the `;`
    disappear, so the next line moves up.
  - `<?= $title; ?>` → `<?= $title ?>`; `<?= $title ;?>` → `<?= $title?>`.

## Options
None.

## PHP versions
None.

## Examples

```php
<?php
declare(ticks=1);

for ($i = 0; $i < 3; $i++) ;
while (next($items)) ;
do ; while (false);
if ($ready) ;
elseif ($late) ;
else ;

function tidy() {
    prepare();<weak_warning descr="Stray semicolon; remove it.">;</weak_warning>
    <weak_warning descr="Stray semicolon; remove it.">;</weak_warning>
    return;
}
class Box {}<weak_warning descr="Stray semicolon; remove it.">;</weak_warning>
?>
<p><?= $title<weak_warning descr="Stray semicolon; remove it.">;</weak_warning> ?></p>
<p><?= $title; $title = null; ?></p>
<p><?php show($title); ?></p>
```

```php
<?php
declare(ticks=1);

for ($i = 0; $i < 3; $i++) ;
while (next($items)) ;
do ; while (false);
if ($ready) ;
elseif ($late) ;
else ;

function tidy() {
    prepare();
    return;
}
class Box {}
?>
<p><?= $title ?></p>
<p><?= $title; $title = null; ?></p>
<p><?php show($title); ?></p>
```

## Divergences
- Alternative syntax (`if (...): ; endif;`, `while (...): ; endwhile;`): the `;`
  sits inside the statement list rather than directly as the clause body;
  upstream behaviour unverified. Recommendation: report it (treat the
  statement list as a block). No fixture covers it.
