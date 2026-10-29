---
id: RandomApiMigration
group: Compatibility
kind: semantic
needs: [names, stubs]
php: { min: "", max: "" }
---

# RandomApiMigration

## Summary

The libc-based `rand()` family is weaker and slower than the Mersenne Twister
family (`mt_*`), and since PHP 7.0 `random_int()` provides a CSPRNG integer.
Suggest renaming legacy random calls to their modern counterpart.

## Detection

Visit every plain function call (not method/static calls).

- **D1** Pick the mapping table:
  - **Modern table** when option `SUGGEST_USING_RANDOM_INT` is on **and** the
    configured PHP language level is **7.0 or higher**:

    | Called    | Suggested      |
    |-----------|----------------|
    | `srand`      | `mt_srand`      |
    | `getrandmax` | `mt_getrandmax` |
    | `rand`       | `random_int`    |
    | `mt_rand`    | `random_int`    |

  - **Classic table** otherwise (option off, or level below 7.0):

    | Called    | Suggested      |
    |-----------|----------------|
    | `srand`      | `mt_srand`      |
    | `getrandmax` | `mt_getrandmax` |
    | `rand`       | `mt_rand`       |

- **D2** The called name — last segment, lower-cased (PHP compares function
  names case-insensitively) — is a key of the chosen table.
- **D3** The call resolves to the **global** built-in function of that name
  (unqualified in the global namespace, `\`-qualified, imported via
  `use function rand;`, or unqualified in a namespace falling back to the
  global one). Calls resolving to a namespaced/user function, unresolvable
  calls, and the name inside a `use function` import line are not reported.
- **D4** Arity adjustment when the suggestion is `random_int` (modern table
  only): `random_int()` requires exactly two arguments, so if the call does
  **not** pass exactly two real arguments (an argument unpacking `...$x` or
  the first-class-callable placeholder `rand(...)` makes the arity unknown,
  whatever the number of entries written):
  - for `rand` the suggestion falls back to `mt_rand`;
  - for `mt_rand` there is no report at all.
- **D5** Report with the (possibly adjusted) suggestion.

## Exceptions (no report)

- **E1** `mt_rand(...)` with an argument count other than 2, or with an
  argument unpacking, under the modern table; `mt_rand` always under the
  classic table.
- **E2** `mt_srand`, `mt_getrandmax`, `random_int` themselves.
- **E3** Calls not resolving to the global function; method/static calls.
- **E4** (custos) Under the modern table, when the file calls `srand()` or
  `mt_srand()` with a seed, `random_int` is not suggested: the sequence is
  meant to be reproducible and `random_int()` ignores the seed. As in D4,
  `rand` then gets the `mt_rand` suggestion and `mt_rand` is not reported.
  `mt_srand()` without an argument reseeds at random and changes nothing.
- **E5** (custos, see Divergences) At PHP 7.1 and above, an `mt_*`
  suggestion (`srand` → `mt_srand`, `getrandmax` → `mt_getrandmax`,
  `rand` → `mt_rand`, whichever table or fallback produced it) is not
  reported: from 7.1 on these names are aliases of their `mt_*`
  counterparts, so the rename changes nothing. `random_int` suggestions
  are still reported.

## Report

- Range: the whole call expression, from the function name (including a
  leading `\` or qualifier as written) to the closing `)`.
- Severity: warning.
- Message: `Prefer {suggested}() over {called}().` — `{called}` is the name as
  written (last segment), `{suggested}` the D1/D4 result.

## Fix

- **F1** Rename to an `mt_*` function: replace only the name identifier (last
  segment) with the suggested name; any leading `\` / qualifier and the
  argument list are kept verbatim.
  - `srand(7)` → `mt_srand(7)`; `\rand()` → `\mt_rand()`;
    `rand($lo, $hi)` → `mt_rand($lo, $hi)` (classic table);
    `rand(5)` → `mt_rand(5)` (modern table, D4 fallback).
- **F2** No fix when the suggestion is `random_int`: the report is kept, the
  code is left alone. `random_int()` draws from the system CSPRNG and ignores
  `srand()` / `mt_srand()`, so a rename would silently break code relying on
  a seeded, reproducible sequence — and the seeding call may live in another
  file. It also rejects `min > max`, which `rand()` tolerates.
- **F3** No fix for `rand(a, b)` → `mt_rand(a, b)` when both bounds are known
  integers (literal, optionally signed, or a resolvable constant) and the
  minimum is above the maximum: `rand(10, 1)` swaps the bounds, `mt_rand(10,
  1)` throws a `ValueError` on PHP 8. Named arguments are matched by name
  (`rand(max: 1, min: 10)`). With bounds that are not both known the rename
  is offered and assumes `min <= max`.
  Builtin spelling: for an unqualified original call the suggested name
  gets a leading `\` when an unqualified call at that position would not reach the global function (a `use function`
  import under that name, or a same-named function declared in the current
  namespace): `namespace Dice; function mt_rand($a, $b) {…}
  rand(1, 6);` → `\mt_rand(1, 6)` (classic table).

## Options

| Option | Type | Default | Effect |
|---|---|---|---|
| `SUGGEST_USING_RANDOM_INT` | bool | `true` | When on and the level is 7.0+, `rand`/`mt_rand` with two arguments are pointed at `random_int`; when off, the classic table is used at every level. |

## PHP versions

- The table choice depends on the level (D1): `>= 7.0` with the option on →
  modern table; otherwise classic.
- From 7.1 on only `random_int` suggestions remain (E5).
- Upstream test cases without an explicit level run at the IDE test default
  (somewhere between 5.6 and 7.0); the only such upstream case turns the
  option off, so the classic table applies regardless. The modern-table case
  sets 7.0 explicitly.

## Examples

Option `SUGGEST_USING_RANDOM_INT = false`, PHP 7.0:

```php
<?php
<warning descr="Prefer mt_srand() over srand().">srand(1234)</warning>;
$cap  = <warning descr="Prefer mt_getrandmax() over getrandmax().">getrandmax()</warning>;
$roll = <warning descr="Prefer mt_rand() over rand().">\rand(1, 6)</warning>;
$keep = mt_rand(1, 6);
```

```php
<?php
mt_srand(1234);
$cap  = mt_getrandmax();
$roll = \mt_rand(1, 6);
$keep = mt_rand(1, 6);
```

Default options, PHP 7.0:

```php
<?php
namespace Dice;

function throwDice($sides)
{
    <warning descr="Prefer mt_srand() over srand().">srand()</warning>;
    $a = <warning descr="Prefer random_int() over rand().">rand(1, $sides)</warning>;
    $b = <warning descr="Prefer random_int() over mt_rand().">mt_rand(0, 9)</warning>;
    $c = <warning descr="Prefer mt_rand() over rand().">rand()</warning>;
    $d = mt_rand();
    $e = mt_rand(3);
    $f = random_int(1, 2);
}
```

```php
<?php
namespace Dice;

function throwDice($sides)
{
    mt_srand();
    $a = rand(1, $sides);
    $b = mt_rand(0, 9);
    $c = mt_rand();
    $d = mt_rand();
    $e = mt_rand(3);
    $f = random_int(1, 2);
}
```

## Divergences

- **Case of the name (custos diverges from upstream).** Upstream compares
  the written name case-sensitively, so `RAND(1, 6)` is not reported. custos
  matches any case; the message names the function in lower case and the fix
  replaces the written name.
- **Argument unpacking (custos diverges).** Upstream counts the entries of
  the argument list, so `mt_rand(...$a, ...$b)` is pointed at `random_int`
  although its real arity is unknown. custos never treats a call containing
  an unpacking as a two-argument call: `rand(...$bounds)` is renamed to
  `mt_rand`, `mt_rand(...$a, ...$b)` is not reported.
- **No automatic rename to `random_int` (custos diverges).** Upstream
  rewrites `rand(a, b)` / `mt_rand(a, b)` to `random_int(a, b)`. That changes
  behavior for seeded generators (F2), so custos only reports.
- **Reversed literal bounds (custos diverges).** Upstream renames
  `rand(10, 1)` to `mt_rand(10, 1)`, which throws on PHP 8; custos reports
  without a fix (F3).
- **Seeded sequences below 7.1.** Before PHP 7.1 `rand()` and `mt_rand()`
  were distinct generators, so renaming seeded `srand($s)` / `rand()` code
  changes the numbers drawn for a given seed there. The `mt_*` renames are
  still offered at those levels, as upstream does.
- **Aliases from 7.1 (custos diverges, E5).** Upstream suggests `mt_rand()`
  over `rand()`, `mt_srand()` over `srand()` and `mt_getrandmax()` over
  `getrandmax()` at every level. From PHP 7.1 on the legacy names are
  aliases of the `mt_*` functions, so the rename is churn with no effect
  (php-src's own test suites call `rand()`); custos reports it below 7.1
  only.
- **Builtin spelling (custos diverges).** Upstream inserts the bare suggested
  name (`mt_srand`, `mt_rand`, `random_int`), which a function of that name declared in (or imported into) the current
  namespace captures, so the fix would call user code. custos writes
  `\name` in that case.
- **Seeded sequences (custos diverges, E4).** Upstream suggests
  `random_int()` for `mt_rand(0, 100)` in a file that calls
  `mt_srand($seed)` to get the same layout from the same seed; following
  that advice breaks the reproducibility the seeding asks for. custos
  suggests nothing for `mt_rand` there and `mt_rand` for `rand`.

Recorded in `internal/testing/conformance/testdata/ea-divergences.json`.
