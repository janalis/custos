---
id: SimpleXmlLoadFileUsage
group: Probable bugs
kind: semantic
needs: [names, stubs]
php: { min: "", max: "" }
---

# SimpleXmlLoadFileUsage

## Summary
`simplexml_load_file()` is affected by a long-standing PHP bug (#62577) where
loading fails depending on the libxml entity-loader state. Reading the file
yourself and passing its contents to `simplexml_load_string()` avoids it.

## Detection
- **D1** A function call (not a method call, not a static call) whose name
  part is `simplexml_load_file`, compared case-insensitively like PHP
  function names (`SimpleXML_Load_File(...)` matches), and which resolves to the global function under PHP's runtime
  rules: `\simplexml_load_file(...)` and an unqualified call in the global
  namespace match; an unqualified call inside a namespace matches only when
  no function of that name is declared in the namespace or imported from
  elsewhere; a qualified `Foo\simplexml_load_file(...)` does not match.
- **D2** The call has at least one argument.

## Exceptions (no report)
- **E1** `simplexml_load_file()` with no arguments.
- **E3** Method calls `$x->simplexml_load_file($f)` / `X::simplexml_load_file($f)`.

## Report
- Range: the whole call expression, from the first character of the name
  (including any namespace qualifier) to the closing `)`.
- Severity: error.
- Message: `simplexml_load_file() is affected by PHP bug #62577; load the contents with file_get_contents() and parse them with simplexml_load_string().`

## Fix
- **F1** Replace the whole call with
  `simplexml_load_string(file_get_contents(A1)REST)` where
  - `A1` is the source text of the first argument, verbatim;
  - `REST` is empty when there is a single argument, otherwise `, ` followed by
    the source texts of the remaining arguments, verbatim, joined with `, `
    (original whitespace/comments between arguments are not kept; the
    arguments' own text is kept as written).
  - `simplexml_load_string` and `file_get_contents` are emitted unqualified
    when a bare call at that position reaches the global function, and with a
    leading `\` otherwise (a function of that name declared in or imported
    into the current namespace).
  Examples: `simplexml_load_file($path)` →
  `simplexml_load_string(file_get_contents($path))`;
  `simplexml_load_file($path,  Node::class,LIBXML_NOCDATA)` →
  `simplexml_load_string(file_get_contents($path), Node::class, LIBXML_NOCDATA)`.

## Options
None.

## PHP versions
No gating upstream; custos reports only below PHP 8.0 (see Divergences).

## Examples

```php
<?php
function feeds(string $dir, string $cls, int $flags, string $nsUri) {
    $empty = simplexml_load_file();
    $a = <error descr="simplexml_load_file() is affected by PHP bug #62577; load the contents with file_get_contents() and parse them with simplexml_load_string().">simplexml_load_file($dir . '/rss.xml')</error>;
    $b = <error descr="simplexml_load_file() is affected by PHP bug #62577; load the contents with file_get_contents() and parse them with simplexml_load_string().">\simplexml_load_file($dir . '/atom.xml', $cls)</error>;
    $c = <error descr="simplexml_load_file() is affected by PHP bug #62577; load the contents with file_get_contents() and parse them with simplexml_load_string().">simplexml_load_file(getenv('FEED'),  $cls,$flags, $nsUri, true)</error>;
    $d = $this->simplexml_load_file('x.xml');
    return [$a, $b, $c, $d, $empty];
}
```

```php
<?php
function feeds(string $dir, string $cls, int $flags, string $nsUri) {
    $empty = simplexml_load_file();
    $a = simplexml_load_string(file_get_contents($dir . '/rss.xml'));
    $b = simplexml_load_string(file_get_contents($dir . '/atom.xml'), $cls);
    $c = simplexml_load_string(file_get_contents(getenv('FEED')), $cls, $flags, $nsUri, true);
    $d = $this->simplexml_load_file('x.xml');
    return [$a, $b, $c, $d, $empty];
}
```

## Divergences
- **Case of the name (custos diverges from upstream).** Upstream matches
  `simplexml_load_file` case-sensitively, so `SimpleXML_Load_File($f)` —
  the same function for PHP — is not reported. custos compares the name
  case-insensitively (D1); the fix emits the lower-case built-in names.
- **Callee resolved — custos diverges from upstream** (D1, F1). Upstream
  matches `Foo\simplexml_load_file(...)` and a namespace-local user function
  of that name, and its fix drops the qualifier, replacing the user's call by
  the built-ins. custos reports only calls reaching the global function, and
  its fix qualifies the generated built-in names with `\` when a bare name
  would resolve to a user function.
- Named arguments (PHP 8): upstream copies argument text as-is, so
  `simplexml_load_file(filename: $f)` would become
  `file_get_contents(filename: $f)` (works by accident). Spread arguments
  (`...$args`) as first argument produce wrong code. Recommendation: skip the
  report when the first argument is a spread or any argument is named.
- **PHP 8.0 and later (custos diverges).** Bug #62577 needs the libxml
  entity loader disabled through `libxml_disable_entity_loader()`, which
  PHP 8.0 deprecated (it requires libxml 2.9, where external entities are
  off by default, so the call is no longer needed). On an 8.0+ target the
  advice only rewrites working code, at error severity (14 findings on
  Akeneo, Mautic and Shopware); custos reports from 5.3 to 7.4 only.
