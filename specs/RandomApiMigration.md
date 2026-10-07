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
  **not** have exactly two arguments:
  - for `rand` the suggestion falls back to `mt_rand`;
  - for `mt_rand` there is no report at all.
- **D5** Report with the (possibly adjusted) suggestion.

## Exceptions (no report)
- **E1** `mt_rand(...)` with an argument count other than 2 under the modern
  table; `mt_rand` always under the classic table.
- **E2** `mt_srand`, `mt_getrandmax`, `random_int` themselves.
- **E3** Calls not resolving to the global function; method/static calls.

## Report
- Range: the whole call expression, from the function name (including a
  leading `\` or qualifier as written) to the closing `)`.
- Severity: warning.
- Message: `Prefer {suggested}() over {called}().` — `{called}` is the name as
  written (last segment), `{suggested}` the D1/D4 result.

## Fix
- **F1** Rename the function: replace only the name identifier (last segment)
  with the suggested name; any leading `\` / qualifier and the argument list
  are kept verbatim.
  - `srand(7)` → `mt_srand(7)`; `\rand()` → `\mt_rand()`;
    `rand($lo, $hi)` → `random_int($lo, $hi)` (modern table);
    `rand(5)` → `mt_rand(5)` (modern table, D4 fallback).
  Builtin spelling: for an unqualified original call the suggested name
  gets a leading `\` when an unqualified call at that position would not reach the global function (a `use function`
  import under that name, or a same-named function declared in the current
  namespace): `namespace Dice; function random_int($a, $b) {…}
  rand(1, 6);` → `\random_int(1, 6)`.

## Options
| Option | Type | Default | Effect |
|---|---|---|---|
| `SUGGEST_USING_RANDOM_INT` | bool | `true` | When on and the level is 7.0+, `rand`/`mt_rand` with two arguments are pointed at `random_int`; when off, the classic table is used at every level. |

## PHP versions
- The table choice depends on the level (D1): `>= 7.0` with the option on →
  modern table; otherwise classic.
- Upstream test cases without an explicit level run at the IDE test default
  (somewhere between 5.6 and 7.0); the only such upstream case turns the
  option off, so the classic table applies regardless. The modern-table case
  sets 7.0 explicitly.

## Examples
Option `SUGGEST_USING_RANDOM_INT = false` (any level):

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
    $a = random_int(1, $sides);
    $b = random_int(0, 9);
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
- `rand(...)` with an argument unpacking (`rand(...$bounds)`) counts as one
  argument and is renamed to `mt_rand`; fine either way. No fixture.
- **Builtin spelling (custos diverges).** Upstream inserts the bare suggested
  name (`mt_srand`, `mt_rand`, `random_int`), which a function of that name declared in (or imported into) the current
  namespace captures, so the fix would call user code. custos writes
  `\name` in that case.
