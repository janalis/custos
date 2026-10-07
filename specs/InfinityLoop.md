---
id: InfinityLoop
group: Probable bugs
kind: syntax
needs: []
php: { min: "", max: "" }
---

# InfinityLoop

## Summary
A method whose only statement calls the very same method on `$this`,
`self` or `static` recurses unconditionally until the stack overflows —
typically a getter/proxy that was meant to call a property, a parent method
or a differently named method.

## Detection
Visit every method declaration (in classes, traits, enums; interface and
abstract methods have no body and are skipped).

- **D1** The method is not abstract and has a body `{ … }`.
- **D2** The body contains exactly **one** statement (comments and doc
  comments are not statements).
- **D3** That statement is either
  - `return <expr>;` → candidate = `<expr>` exactly as written (a
    parenthesized expression `return ($this->m());` is *not* unwrapped and
    therefore does not qualify), or
  - an expression statement `<expr>;` → candidate = `<expr>`.
  Any other statement kind (`echo`, `if`, assignments as the expression,
  `return;`) → stop.
- **D4** The candidate is a method call (`->`, `?->` or `::`) whose method
  name as written equals the declaring method's name ignoring case (PHP
  method names are case-insensitive, so `$this->Load()` inside `load()`
  calls itself).
- **D5** The receiver/class part of the call, as source text, is `$this`,
  `self` or `static`; the keywords `self`/`static` are matched ignoring case
  (`SELF::`, `Static::`), `$this` exactly (variable names are
  case-sensitive). `parent`, other variables and class names do not
  qualify. Arguments are irrelevant.

## Exceptions (no report)
- **E1** Methods with more than one statement (even if the recursive call
  is guarded or not).
- **E2** `parent::sameName()`, `ClassName::sameName()`, `$other->sameName()`.
- **E3** `$This->m()` (a different variable).
- **E4** Recursive call nested in an expression (`return 1 + $this->m();`,
  `$x = $this->m();`, `return ($this->m());`).
- **E5** Plain functions and closures.

## Report
- Range: the method call expression, from `$this`/`self`/`static` to the
  closing `)` (no `return` keyword, no `;`).
- Severity: error.
- Message: `Method calls itself unconditionally; this recursion never ends.`

## Fix
None.

## Options
None.

## PHP versions
None.

## Examples

```php
<?php
class Settings
{
    private $theme;

    public function theme()
    {
        return <error descr="Method calls itself unconditionally; this recursion never ends.">$this->theme()</error>;
    }

    public static function boot($env)
    {
        <error descr="Method calls itself unconditionally; this recursion never ends.">static::boot($env)</error>;
    }

    protected function flush()
    {
        // delegate
        <error descr="Method calls itself unconditionally; this recursion never ends.">self::flush()</error>;
    }

    public function locale()
    {
        return $this->theme;
    }

    public function reload()
    {
        return parent::reload();
    }

    public function save()
    {
        return <error descr="Method calls itself unconditionally; this recursion never ends.">$this->Save()</error>;
    }

    public function sync()
    {
        return ($this->sync());
    }

    public function close($n)
    {
        if ($n > 0) {
            return $this->close($n - 1);
        }
    }
}
```

## Divergences
- **Case-insensitive keywords and names (custos diverges):** upstream
  compares the receiver text and the method name literally, so
  `SELF::flush()`, `Static::boot()` or `$this->Save()` inside `save()` are
  not reported, although PHP resolves keywords and method names
  case-insensitively and each of these recurses forever. custos ignores case
  for the `self`/`static` keywords and the method name (D4, D5).
