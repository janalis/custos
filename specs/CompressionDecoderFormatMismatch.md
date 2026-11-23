---
id: CompressionDecoderFormatMismatch
group: Probable bugs
kind: flow
needs: [names, types, flow]
php: { min: "", max: "" }
---

# CompressionDecoderFormatMismatch

## Summary

Match the decoder to the compression format.

## Detection

- D1. Compare resolved gzip/zlib/raw encoders and decoders through direct nesting or unchanged locals. Report gzencode input to gzuncompress/gzinflate, gzcompress input to gzdecode/gzinflate, or gzdeflate input to gzdecode/gzuncompress. Encoders with a third explicit format argument and manually transformed bytes remain unknown.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape and unresolved calls that can mutate relevant state remove proof. When the proof is inconclusive, nothing is reported.

## Exceptions (no report)

An explicit encoding argument stays unknown whether positional or named. An encoder called with unpacked arguments (`gzencode(...$args)`) is unknown too, because the spread may carry the encoding. A decoded local is known only while nothing in its scope can change it behind its plain assignments, anywhere and in any statement order: `++` or `--` on it or on one of its elements; a `catch`, `foreach` key or value binding it; a reference to it (a by-reference assignment on either side, a by-reference `foreach` over it, a by-reference array item, a by-reference capture by a nested closure, or an argument, other than the inspected argument itself, that a resolved callee binds by reference or that a callee custos cannot resolve receives); a write naming its target at run time (a variable variable, `extract()`, one-argument `parse_str()` or `mb_parse_str()`, `include`/`require`, `eval`); or a `$GLOBALS` access other than a literal entry naming another variable. A `global` or `static` declaration of it, or a by-reference capture by the closure that is the scope, removes the proof too. In top-level code, where locals are globals, a `global` declaration naming it in any function of the file, or a `$GLOBALS` access anywhere in the file, does as well. The same holds for a local copied into it by a plain assignment (`$copy = $bytes;`), directly or through further copies; the operands of a computed value do not count, since the proof is about where the bytes come from. A read inside an arrow function also checks the scopes the arrow function captures from. The replacement decoder must exist in the target version; gzdecode is unavailable on PHP 5.3, so that finding has no automatic fix.

- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding. Both legacy noinspection and custos-ignore forms suppress this ID.

## Report

- Range: the violating expression or call described in D1; stateful absence checks highlight the final unsafe operation or loop.
- Severity: warning.
- Enabled by default: true.
- Message: `Match the decoder to the compression format.`

## Fix

Replace only the resolved decoder function name: gzencode→gzdecode, gzcompress→gzuncompress, gzdeflate→gzinflate. Preserve namespace syntax; omit if a matching replacement builtin is shadowed.

All edits use exact byte ranges and preserve comments and evaluation order. Offer no fix when its stated prerequisites cannot be proven or the replacement is unsupported by the configured PHP version.

## Options

None.

## PHP versions

Requires PHP 5.3 or later and the referenced extension API. Respect method-specific introduction versions within the supported PHP versions; unsupported optional APIs cannot establish facts. Named arguments require PHP 8.0 or later.

## Examples

```php
<?php
<warning descr="Match the decoder to the compression format.">gzuncompress(gzencode("sample payload"))</warning>;
```

Fixed result:

```php
<?php
gzdecode(gzencode("sample payload"));
```

Valid case:

```php
<?php
gzdecode(gzencode("sample payload"));
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
