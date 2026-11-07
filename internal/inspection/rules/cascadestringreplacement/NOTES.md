# CascadeStringReplacement implementation notes

## Fix prerequisites

`fixable` checks F0's known-shape and pairing conditions over the calls,
latest first. Merged lists hold the earliest call's elements first, so every
call but the latest must contribute as many replacements as searches:

- an argument that may be a string or an array (untyped, `array|string`)
  cannot be written as one element nor spread;
- a string search needs a string replace;
- a string replace is copied once per search element, which needs their
  number (and at least one);
- array search and replace must have the same known length, except in the
  latest call, whose elements end both lists (PHP pads a short replace list
  with `''` and ignores extra entries).

When every call replaces with the same string literal, the replace stays one
scalar and only the searches are merged. Without that shortcut, EspoCRM's
sheet-name cleanup turned `:` and `/` into `''` instead of `' '`.

## Empty lists and cardinality

Empty search lists contribute no replacement pairs. Before merging, empty
replacement lists are padded with explicit empty strings so later pairs cannot
shift into them. Dropped evaluations must be harmless; scalar coercions,
assignments, calls, references and uncertain keys keep the fix unavailable.
Exact native array cardinality can expand a scalar replacement for a known
number of searches. PHPDoc-only shapes do not establish that count.
