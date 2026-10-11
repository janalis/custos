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
D2. The method's modifier list holds at least **two** modifier keywords.
    Only keyword tokens count: whitespace and comments between keywords are
    never read as modifiers, whatever words a comment contains.
D3. Build *expected* by stably sorting the keywords as written: `final`
    first, then `abstract`, then the visibility keyword, then `static`.
    Sorting only moves keywords; it never adds or removes one.
D4. Report when the keyword sequence differs from *expected*.
    Consequences: line breaks, extra spaces and comments between correctly
    ordered keywords are **not** reported; upper/mixed-case keywords in the
    correct order are not reported.

## Exceptions (no report)

E1. Methods with none of `static`/`abstract`/`final` (e.g. only `public`).
E2. Methods with a single modifier keyword.
E3. Already-canonical order, regardless of spacing/line breaks/case:
    `final public static`, `abstract protected`, `final static`,
    `public\n    static`, `PUBLIC STATIC`,
    `public /* not final */ static`.
E4. Properties, constants, class declarations, promoted parameters — only
    methods are inspected.

## Report

- Range: the modifier list, from the start of its first keyword to the end of
  its last keyword (interior whitespace/newlines included).
- Severity: info (fixtures tag it `weak_warning`).
- Message: `Reorder modifiers as: {expected}.` with the sorted keywords
  lowercased and joined by single spaces.

## Fix

F1. Without a comment in the list: replace the reported range with the
    lowercased sorted keywords separated by single spaces. Multi-line lists
    collapse to one line; text after the list (` function name(...)`) is
    untouched.
F2. With a comment between keywords: write the lowercased sorted keywords into
    the keyword positions and keep everything between them (comments, spaces,
    line breaks) byte for byte, so
    `static /* was protected */ public` becomes
    `public /* was protected */ static`.

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

- Comments between modifiers are ignored when deciding the order, and the fix
  keeps them in place. Upstream compares the raw text of the list, so it
  reports a correctly ordered list that merely contains a comment, treats
  keyword-like words inside a comment as modifiers, and its fix deletes the
  comment. Following that would let the fix add `final` or a second visibility
  the method never had (no upstream fixture covers comments).
