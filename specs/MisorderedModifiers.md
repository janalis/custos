---
id: MisorderedModifiers
group: Code style
kind: syntax
needs: []
php: { min: "", max: "" }
---

# MisorderedModifiers

## Summary
PSR-1/PSR-12 style expects method modifiers in a fixed order: `final` /
`abstract` first, then the visibility, then `static`. Methods whose modifiers
are written in another order are flagged and re-sorted.

## Detection
D1. Node: a method declaration (in any class-like: class, abstract class,
    interface, trait, enum, anonymous class) that carries at least one of the
    modifiers `static`, `abstract`, `final`.
D2. The method's modifier list consists of at least **two** non-whitespace
    tokens (keywords; a comment placed between keywords also counts as a
    token).
D3. Build *original*: the modifier list's non-whitespace tokens, in source
    order, joined by single spaces, lowercased.
    Build *expected*: walk the canonical sequence
    `final`, `abstract`, `public`, `protected`, `private`, `static` and keep
    each keyword that occurs as a substring of *original*; join the kept ones
    with single spaces.
D4. Report when *original* ≠ *expected*.
    Consequences: line breaks / extra spaces between correctly ordered
    keywords are **not** reported (whitespace is normalised in *original*);
    upper/mixed-case keywords in the correct order are not reported; a comment
    inside the modifier list always makes the strings differ (reported).

## Exceptions (no report)
E1. Methods with none of `static`/`abstract`/`final` (e.g. only `public`).
E2. Methods with a single modifier keyword.
E3. Already-canonical order, regardless of spacing/line breaks/case:
    `final public static`, `abstract protected`, `final static`,
    `public\n    static`, `PUBLIC STATIC`.
E4. Properties, constants, class declarations, promoted parameters — only
    methods are inspected.

## Report
- Range: the modifier list, from the start of its first keyword to the end of
  its last keyword (interior whitespace/newlines included).
- Severity: info (fixtures tag it `weak_warning`).
- Message: `Reorder modifiers as: {expected}.`

## Fix
F1. Replace the reported range with *expected* (lowercase keywords separated by
    single spaces). Multi-line lists collapse to one line; comments inside the
    list are dropped; text after the list (` function name(...)`) is untouched.

## Options
| Option | Type | Default | Effect |
|---|---|---|---|

## PHP versions
None.

## Examples
```php
<?php

interface Registry
{
    <weak_warning descr="Reorder modifiers as: public static.">static public</weak_warning> function instance();
    public static function reset();
}

abstract class Repository
{
    public static function boot() {}
    abstract protected function table();
    final static function hash() {}
    protected
        static function cache() {}
    PRIVATE STATIC function lower() {}

    <weak_warning descr="Reorder modifiers as: abstract protected.">protected abstract</weak_warning> function columns();
    <weak_warning descr="Reorder modifiers as: final protected static.">static
        protected final</weak_warning> function guard() {}
    <weak_warning descr="Reorder modifiers as: final private static.">Static Private Final</weak_warning> function seal() {}
    <weak_warning descr="Reorder modifiers as: abstract public static.">public static abstract</weak_warning> function make();
}
```

```php
<?php

interface Registry
{
    public static function instance();
    public static function reset();
}

abstract class Repository
{
    public static function boot() {}
    abstract protected function table();
    final static function hash() {}
    protected
        static function cache() {}
    PRIVATE STATIC function lower() {}

    abstract protected function columns();
    final protected static function guard() {}
    final private static function seal() {}
    abstract public static function make();
}
```

## Divergences
- A comment between modifiers (`static /* x */ public`) is reported upstream
  and the fix deletes the comment; even a correctly ordered list with a comment
  is reported. Recommendation: ignore comment tokens when building *original*
  and preserve them otherwise (no fixture covers comments).
