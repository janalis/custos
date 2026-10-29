---
id: ClassConstantCanBeUsed
group: Language level migration
kind: semantic
needs: [names, index, stubs]
php: { min: "5.5", max: "" }
---

# ClassConstantCanBeUsed

## Summary

Since PHP 5.5 a class name can be obtained with `Name::class`. Writing class
names as string literals hides them from refactoring tools and static
analysis, and `get_called_class()` / `get_parent_class()` without arguments
have constant-expression equivalents (`static::class`, `parent::class`).

## Detection

Everything is gated on PHP language level ≥ 5.5.

### Function calls

- **D0** *Class scope* of a call: walk up from the call; a named function
  declaration (`function f() {}`, even one written inside a method) ends the
  walk with no class scope; otherwise the nearest enclosing class-like
  (class, anonymous class, interface, trait, enum) is the class scope.
  Closures and arrow functions are transparent.
- **D1** A function call (not a method call) to the global
  `get_called_class` (name compared case-insensitively; written unqualified
  or `\`-qualified, not shadowed by a function of that name declared in or
  imported into the current namespace) with no arguments, and the call has a
  class scope (D0) → suggest `static::class`.
- **D2** A function call to the global `get_parent_class` (same matching as
  D1) with no arguments, whose class scope (D0) is a class
  (named or anonymous, not a trait/interface/enum) that declares `extends`
  → suggest `parent::class`.

### String literals

- **D3** A single- or double-quoted string literal (not heredoc/nowdoc) with
  no interpolation (no `$var`, `{$…}` parts). A double-quoted literal is
  skipped when it holds an escape sequence that PHP decodes: a backslash
  followed by `n`, `t`, `r`, `v`, `e`, `f`, an octal digit, or `x` and a
  hexadecimal digit. A doubled backslash is a plain backslash and is read
  as a pair, so `"Shop\\name"` is still examined while `"Shop\name"` is not.
- **D4** Context filter on the literal's direct parent:
  - if the parent is a binary expression, it is processed only when the
    operator is concatenation `.` and the *left* operand is the constant
    `__NAMESPACE__` (so the literal is the right operand) — the *namespace
    concat* case. Any other binary parent (`'X' . $y`, `$y === 'X'`, …) → skip;
  - if the parent is a compound assignment (`.=`, `+=`, `??=`, …) → skip;
  - if the literal is the 2nd argument of a function call named
    `class_alias` (any case, resolving to the global function) that has
    exactly 2 arguments → skip.
  Any other context (argument, array element/key, return value, assignment
  right side, parenthesised, …) is processed.
- **D5** Let `raw` be the literal's raw contents (text between the quotes).
  For a plain literal the candidate is `raw`. For the namespace-concat case
  `raw` must start with a backslash, otherwise there is no report; the
  candidate is then `\` + the name of the namespace block enclosing the
  literal + `raw`, or just `raw` when that namespace is the global one.
- **D6** The candidate must be longer than 3 characters and consist only of
  ASCII letters, digits and the characters `_ \ [ ] ^` and backtick (upstream
  uses a character range that accidentally admits those punctuation
  characters; reproduce it).
- **D7** If the candidate contains no backslash and equals its own lowercase
  form, skip (e.g. `'stdclass'`, `'name'`).
- **D8** Normalize: replace every doubled backslash `\\` by a single `\`
  (left to right, non-overlapping). Then:
  - if it starts with `\`, it is the FQN to look up;
  - else if it contains a `\`, or option `LOOK_ROOT_NS_UP` is on, the FQN is
    `\` + normalized text (i.e. relative-looking names are always treated as
    fully qualified from the root, never resolved against the current
    namespace or imports);
  - else skip.
- **D9** Look up the FQN among all classes and interfaces (and enums) known to
  the project/index and PHP stubs (traits are not considered). Report only if
  exactly one declaration is found **and** its FQN matches the looked-up FQN
  exactly, including letter case. Duplicate declarations of the same FQN → no
  report.

## Exceptions (no report)

- **E1** PHP level below 5.5.
- **E2** `get_called_class(...)` / `get_parent_class(...)` with any argument.
- **E7** `get_called_class()` without a class scope (file level, plain
  functions); `get_parent_class()` outside a class that declares `extends`
  (file level, plain functions, parentless classes, traits).
- **E3** Interpolated strings, heredoc/nowdoc, double-quoted literals with a
  decoded escape sequence (`"app\events\Signal"` holds an ESC byte).
- **E4** Literals that are operands of a binary expression other than
  `__NAMESPACE__ . 'literal'`; literals on the right of a compound assignment;
  the alias argument of a 2-argument `class_alias()`.
- **E5** Candidates of length ≤ 3, containing other characters (spaces, `-`,
  `:`, `.`, `$`, …), lowercase-only without backslash, or unqualified when
  `LOOK_ROOT_NS_UP` is off.
- **E8** `__NAMESPACE__ . 'Name'` without a backslash at the start of the
  literal: the runtime value is `NsName`.
- **E6** Names that do not resolve, resolve to a trait, resolve ambiguously,
  or differ in case from the declaration.

## Report

- Range:
  - D1/D2: the whole call, from the function name to `)`.
  - plain literal: the string literal including its quotes;
  - namespace concat: the whole binary expression `__NAMESPACE__ . '…'`.
- Severity: info (weak warning).
- Message:
  - D1: `Use static::class instead.`
  - D2: `Use parent::class instead.`
  - literals: `Use {N}::class instead of the class name string.` where `{N}`
    is: for a plain literal, the normalized FQN with leading `\`; for the
    namespace concat case, the literal's raw contents with `\\` collapsed to
    `\` and one leading `\` removed (a namespace-relative name, e.g.
    `Sub\Item`).

## Fix

- **F1** D1: replace the call with `static::class`.
- **F2** D2: replace the call with `parent::class`.
- **F0** No fix (report only, with the leading-backslash message) when the
  string's value starts with a backslash: a plain literal written with a
  leading `\`, or `__NAMESPACE__ . '\Name'` in the global namespace.
- **F3** Plain literals: the fix receives `Q`, the normalized FQN with
  leading `\`, and computes the replacement class reference `T`; the reported
  range is replaced with `T::class`. Let `short` be the part of `Q` after its
  last `\`, and *the block* the namespace block that contains the literal
  (each `namespace` declaration starts a new block, even when it repeats a
  name; a file without any declaration is one block). PHP resets the import
  table at every namespace declaration, so nothing outside the block counts.
  1. Scan the class imports of the block in document order (`use function`
     and `use const` items are ignored; they live in other symbol tables).
     The last `use` statement of the block, of any kind, is the import
     marker. For each imported class:
     - its fully-qualified name equals `Q` with the same letter case: `T` =
       the item's local name (alias if aliased, otherwise last segment),
       *already imported* (the last match wins);
     - else its fully-qualified name equals `Q` ignoring case, or its local
       name equals `short` ignoring case: *no import*. `T` stays `Q`.
  2. If not already imported, `T` = `Q` initially.
  3. If option `IMPORT_CLASSES_ON_QF` is on, the scan allowed it, `Q`
     contains at least 2 backslashes, and no class-like named `short` exists
     in the block's namespace (project index; any kind; the global namespace
     included):
     - if `USE_RELATIVE_QF` is on, the block has a named namespace `NS` and
       `Q` starts with `\NS\`: `T` = `Q` without that prefix (e.g.
       `Sub\Item`), unless a class import of the block has the first segment
       of that relative name as local name (it would capture it); then `T`
       stays `Q`. No import is added either way.
     - otherwise, unless the block declares a class-like named `short` or
       references a name whose first segment is `short` (unqualified or
       qualified names as written, plus relative names written by earlier
       fixes; ignoring case): add an import and set `T` = `short`.
       The import `use <Q without leading \>;` goes
       - after the marker (as a new line following the last `use` statement
         of the block), or, without a marker,
       - before the first statement of the block that is not inline HTML,
         followed by a blank line; when that statement is a `declare(...)`,
         before the statement following it. If there is none, no import is
         added and `T` stays `Q`.
       A file without namespace declaration always uses the second form,
       even when it has `use` statements.
  4. Replace the reported range with `T::class`.
- **F4** Namespace concat: let `R` be the literal's contents with `\\`
  collapsed and the leading `\` removed. The expression is replaced with
  `R::class`, or with `namespace\R::class` when a class import of the block
  has the first segment of `R` as local name. No import is ever added.
- Multiple fixes in one file are applied in document order; imports added by
  earlier fixes are seen by later ones in the same block (a later literal
  for the same class then reuses the import; a later class with the same
  short name then stays fully qualified). Every class that may be imported
  gets its own `use` line; a fix whose import cannot be written next to the
  imports of the fixes applied with it in the same pass is left for the next
  pass, where it is placed after them. Applied alone, each fix is complete;
  a single pass over many classes therefore leaves some literals for the
  next one instead of writing them fully qualified.
- Results:
  - plain literal, imported via `use Ns\Thing as Other;` → `Other::class`;
  - imported with another letter case (`use ns\thing;`) → `\Ns\Thing::class`;
  - root-namespace class (`\Thing`, only one backslash) never gets an import
    → `\Thing::class` (unless imported/aliased already);
  - collision with an existing import's local name, with a class-like of the
    block's namespace, or with a name the block already uses →
    `\Full\Name::class`;
  - namespace concat `__NAMESPACE__ . '\Item'` → `Item::class`;
    `__NAMESPACE__ . '\Sub\Item'` → `Sub\Item::class`; with `use Other\Item;`
    in the block → `namespace\Item::class`.

## Options

| Option | Type | Default | Effect |
|---|---|---|---|
| `IMPORT_CLASSES_ON_QF` | bool | true | Fix may add a `use` import for namespaced classes and reference them by short name. |
| `USE_RELATIVE_QF` | bool | false | When importing is otherwise possible, a class below the enclosing namespace is written as a namespace-relative name instead of being imported. |
| `LOOK_ROOT_NS_UP` | bool | false | Also consider unqualified strings (no backslash), looked up as root-namespace classes (`'Thing'` → `\Thing`). |

## PHP versions

- Whole rule requires PHP ≥ 5.5. Upstream tests run at the test default
  level (below 7.1, ≥ 5.5).

## Examples

Options: `IMPORT_CLASSES_ON_QF=true`, `USE_RELATIVE_QF=true`,
`LOOK_ROOT_NS_UP=true`.

```php
<?php

namespace Shop\Billing\Tax;
class Rate {}

namespace Vendor\Kit;
class Ledger {}
class Meter {}

namespace Shop\Billing;

use Shop\Billing\Invoice as Bill;
use \ArrayObject as Bag;

class Invoice {}
class Ledger {}

class Clerk {
    public function names() {
        return [
            <weak_warning descr="Use static::class instead.">get_called_class()</weak_warning>,
            get_parent_class(),
            <weak_warning descr="Use \ArrayObject::class instead of the class name string (::class has no leading backslash).">'\ArrayObject'</weak_warning>,
            <weak_warning descr="Use \ArrayObject::class instead of the class name string.">'ArrayObject'</weak_warning>,
            <weak_warning descr="Use \Shop\Billing\Invoice::class instead of the class name string.">'Shop\Billing\Invoice'</weak_warning>,
            <weak_warning descr="Use \Shop\Billing\Ledger::class instead of the class name string.">"Shop\\Billing\\Ledger"</weak_warning>,
            <weak_warning descr="Use Ledger::class instead of the class name string.">__NAMESPACE__ . '\Ledger'</weak_warning>,
            <weak_warning descr="Use \Shop\Billing\Tax\Rate::class instead of the class name string.">'Shop\Billing\Tax\Rate'</weak_warning>,
            <weak_warning descr="Use \Vendor\Kit\Ledger::class instead of the class name string.">'Vendor\Kit\Ledger'</weak_warning>,
            <weak_warning descr="Use \Vendor\Kit\Meter::class instead of the class name string.">'Vendor\Kit\Meter'</weak_warning>,
            <weak_warning descr="Use \Vendor\Kit\Meter::class instead of the class name string (::class has no leading backslash).">'\Vendor\Kit\Meter'</weak_warning>,
        ];
    }

    public function untouched($suffix) {
        $s  = '';
        $s .= '\ArrayObject';
        class_alias($suffix, '\Shop\Billing\Ledger');
        return [
            get_called_class($this),
            'arrayobject',
            'Bag',
            'Invoice',
            'shop\billing\invoice',
            '\ArrayObject' . $suffix,
            "\\ArrayObject{$suffix}",
            $suffix == 'Shop\Billing\Ledger',
            'Shop\Billing\Missing',
            'Not a class',
        ];
    }
}
```

```php
<?php

namespace Shop\Billing\Tax;
class Rate {}

namespace Vendor\Kit;
class Ledger {}
class Meter {}

namespace Shop\Billing;

use Shop\Billing\Invoice as Bill;
use \ArrayObject as Bag;
use Vendor\Kit\Meter;

class Invoice {}
class Ledger {}

class Clerk {
    public function names() {
        return [
            static::class,
            get_parent_class(),
            '\ArrayObject',
            Bag::class,
            Bill::class,
            \Shop\Billing\Ledger::class,
            Ledger::class,
            Tax\Rate::class,
            \Vendor\Kit\Ledger::class,
            Meter::class,
            '\Vendor\Kit\Meter',
        ];
    }

    public function untouched($suffix) {
        $s  = '';
        $s .= '\ArrayObject';
        class_alias($suffix, '\Shop\Billing\Ledger');
        return [
            get_called_class($this),
            'arrayobject',
            'Bag',
            'Invoice',
            'shop\billing\invoice',
            '\ArrayObject' . $suffix,
            "\\ArrayObject{$suffix}",
            $suffix == 'Shop\Billing\Ledger',
            'Shop\Billing\Missing',
            'Not a class',
        ];
    }
}
```

Walk-through of the less obvious results:

- `'\ArrayObject'` and `'\Vendor\Kit\Meter'`: the value starts with a
  backslash → reported without a fix (F0).
- `"Shop\\Billing\\Ledger"`: the class lives directly in the enclosing
  namespace, so the namespace-collision check (`Shop\Billing\Ledger` exists)
  finds the class itself → no import, stays fully qualified.
- `'Shop\Billing\Tax\Rate'`: below the enclosing namespace, no collision →
  relative name `Tax\Rate` (USE_RELATIVE_QF), no import.
- `'Vendor\Kit\Ledger'`: `Shop\Billing\Ledger` exists → stays fully qualified.
- `'Vendor\Kit\Meter'`: imported after the last `use` of the block.
- `'shop\billing\invoice'`: resolves only case-insensitively → not reported.
- `'Bag'`: only 3 characters.
- `get_parent_class()`: `Clerk` has no parent, so `parent::class` would be
  a fatal error → not reported (E7).

No-namespace file with `declare` (options `true, false, false`):

```php
<?php

declare(strict_types=1);

namespace_free();

final class Local {}

return [
    <weak_warning descr="Use \Shop\Billing\Invoice::class instead of the class name string.">'Shop\Billing\Invoice'</weak_warning>,
    <weak_warning descr="Use \Shop\Billing\Invoice::class instead of the class name string (::class has no leading backslash).">'\Shop\Billing\Invoice'</weak_warning>,
    'Local',
];
```

(with `Shop\Billing\Invoice` declared in another project file)

```php
<?php

declare(strict_types=1);

use Shop\Billing\Invoice;

namespace_free();

final class Local {}

return [
    Invoice::class,
    '\Shop\Billing\Invoice',
    'Local',
];
```

## Divergences

- **Namespace concat (custos diverges).** Upstream resolves
  `__NAMESPACE__ . '\Name'` through the enclosing class (root namespace
  outside a class), accepts `__NAMESPACE__ . 'Name'` although it evaluates
  to `NsName`, and sends the relative name through the import logic, which
  can add a bogus `use A\B\C;` and lets an existing `use Other\Name;`
  capture the result. custos resolves against the enclosing namespace
  block, requires the leading backslash (D5, E8), never imports and writes
  `namespace\Name::class` when an import would capture the name (F4). In
  the global namespace the value is `'\Name'`, so there is no fix (F0).
- **Escape sequences (custos diverges).** Upstream examines double-quoted
  literals raw, so `"app\events\Signal"` is matched as a class name although
  `\e` is the ESC byte at runtime. custos skips literals with a decoded
  escape sequence (D3).
- **Namespace blocks (custos diverges).** Upstream scans the imports of the
  whole file and uses the file's *first* namespace for the collision check,
  the relative name and the insertion point. In a file with several
  namespace blocks that reuses an alias that means something else where the
  literal is, or imports into a block the literal is not in. custos scopes
  imports, collision checks and the insertion point to the block enclosing
  the literal (F3).
- **Import collisions (custos diverges).** Upstream only keeps the fully
  qualified name when a *class* of that short name exists in the first
  namespace. custos also keeps it for interfaces, traits and enums, in the
  global namespace, when the block declares the short name itself and when
  the block already references it: the new import would be a "cannot
  redeclare" fatal error or would change what those references mean. With
  `USE_RELATIVE_QF`, a relative name whose first segment is an imported
  alias is not written either.
- **Import kinds and case (custos diverges).** Upstream treats `use
  function` / `use const` items like class imports and custos used to reuse
  a class import that differs only in letter case. `X::class` returns the
  spelling of the `use` statement, and a function import does not name a
  class, so both changed the string. custos reuses class imports with the
  exact case only; a differently-cased one blocks a second import and keeps
  the fully qualified name.
- custos diverges from upstream on the class scope of D1/D2 (D0, E7).
  Upstream suggests `static::class` and `parent::class` wherever the calls
  appear. Outside a class scope `static::class` fails, and in a class
  without a parent `parent::class` is a compile-time error while
  `get_parent_class()` simply returns `false`; in a trait it depends on the
  using class. custos only reports `get_called_class()` with a class scope
  and `get_parent_class()` in a class that declares `extends`. The EA case
  whose parentless class expects the `parent::class` rewrite is listed in
  `testdata/ea-divergences.json`.
- **Function names (custos diverges):** upstream matches
  `get_called_class` / `get_parent_class` / `class_alias` on the written last
  segment, case-sensitively: it misses `Get_Parent_Class()` and rewrites a
  user `App\get_called_class()` to `static::class`. custos matches any case and
  requires the call to resolve to the global function (D1, D2, D4).
- **Local name case (custos diverges):** upstream compares local import
  names case-sensitively, so with `use Other\LEDGER;` the fix for
  `'Vendor\Kit\Ledger'` adds `use Vendor\Kit\Ledger;`, a fatal "name already
  in use" error. custos compares local names ignoring case (F3 step 1).
- **Leading backslash (custos diverges).** `X::class` never starts with a
  backslash, so rewriting `'\Shop\Jar'` to `\Shop\Jar::class` changes the
  string's value (`'\Shop\Jar' === 'Shop\Jar'` is false): string
  comparisons, array keys and maps built from such literals behave
  differently after the fix. custos still reports these literals but
  without a fix, with the message `Use {name}::class instead of the class
  name string (::class has no leading backslash).` Literals without a
  leading backslash (and `__NAMESPACE__ . '\X'` in a named namespace, whose
  value has none) keep the fix. EA cases affected: `class-name-constant-collisions.php`,
  `configuration-class-reference.php` (listed divergences).
- Conformance note: upstream fixtures run with `IMPORT_CLASSES_ON_QF=true,
  USE_RELATIVE_QF=true, LOOK_ROOT_NS_UP=true`, except the no-namespace
  configuration case (`true, false, false`). Some upstream cases rely on
  classes declared in a companion file (project index).
