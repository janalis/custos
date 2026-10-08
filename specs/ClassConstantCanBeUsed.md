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
  no interpolation (no `$var`, `{$…}` parts).
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
- **D5** Let `raw` be the literal's raw contents (text between the quotes,
  escape sequences *not* decoded). For the namespace-concat case, if the
  literal is inside a class-like declaration (class, interface, trait, enum —
  nearest enclosing), the candidate is: that class's namespace (fully
  qualified) + `\` + `raw` with one leading backslash removed. If there is no
  enclosing class-like, the candidate is just `raw` (see Divergences).
  Otherwise (plain literal) the candidate is `raw`.
- **D6** The candidate must be longer than 3 characters and consist only of
  ASCII letters, digits and the characters `_ \ [ ] ^` and backtick (upstream
  uses a character range that accidentally admits those punctuation
  characters; reproduce it).
- **D7** If the candidate contains no backslash and equals its own lowercase
  form, skip (e.g. `'stdclass'`, `'name'`).
- **D8** Normalise: replace every doubled backslash `\\` by a single `\`
  (left to right, non-overlapping). Then:
  - if it starts with `\`, it is the FQN to look up;
  - else if it contains a `\`, or option `LOOK_ROOT_NS_UP` is on, the FQN is
    `\` + normalised text (i.e. relative-looking names are always treated as
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
- **E3** Interpolated strings, heredoc/nowdoc.
- **E4** Literals that are operands of a binary expression other than
  `__NAMESPACE__ . 'literal'`; literals on the right of a compound assignment;
  the alias argument of a 2-argument `class_alias()`.
- **E5** Candidates of length ≤ 3, containing other characters (spaces, `-`,
  `:`, `.`, `$`, …), lowercase-only without backslash, or unqualified when
  `LOOK_ROOT_NS_UP` is off.
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
    is: for a plain literal, the normalised FQN with leading `\`; for the
    namespace concat case, the literal's raw contents with `\\` collapsed to
    `\` and one leading `\` removed (a namespace-relative name, e.g.
    `Sub\Item`).

## Fix

- **F1** D1: replace the call with `static::class`.
- **F2** D2: replace the call with `parent::class`.
- **F3** Literals: the fix receives a name `Q` — the normalised FQN with
  leading `\` for plain literals, or the relative name of the message for the
  namespace-concat case — and computes the replacement class reference `T`;
  the reported range is replaced with `T::class`.
  Let `short` = the part of `Q` after its last `\`.
  1. Scan every `use` import statement of the file in document order (all
     namespaces of the file; class, function and const imports alike;
     excluding closure `use (...)` clauses and trait `use` inside class
     bodies). Remember the *last* such statement as the import marker. For
     each imported item in each statement:
     - if the item's fully-qualified imported name (with leading `\`) equals
       `Q` ignoring case (class names are case-insensitive): set `T` = the item's local name (alias if aliased,
       otherwise last segment), mark *already imported*, and stop scanning
       that statement (later statements are still scanned and a later match
       overwrites `T` — the last match wins);
     - else if the item's local name equals `short` ignoring case (PHP
       rejects `use A\Ledger; use B\LEDGER;`): mark *name collision*.
  2. If not already imported, `T` = `Q` initially.
  3. If option `IMPORT_CLASSES_ON_QF` is on, not already imported, no name
     collision, and `Q` contains at least 2 backslashes:
     - Let `NS` be the **first** namespace declaration of the file (not
       necessarily the enclosing one).
     - With `NS`: if a class (classes only) `NS\short` exists in the project,
       do nothing more (`T` stays `Q`, fully qualified). Otherwise:
       - if there is no import marker, the marker becomes the first statement
         inside `NS` and the import is inserted *before* it;
       - if `USE_RELATIVE_QF` is on and `Q` starts with `\` + `NS` + `\`,
         set `T` = `Q` with that prefix removed (e.g. `Sub\Item`) and do not
         add any import.
     - Without any namespace: the marker becomes the first statement of the
       file, inserted *before* it (this replaces a marker found in step 1 —
       upstream behaviour).
     - If an import is still to be added: when the marker is a `declare(...)`
       statement, insert before the statement following it instead (so the
       import goes after `declare(strict_types=1);`). Insert
       `use <Q without leading \>;`
       - after the marker (as a new line following the last existing `use`
         statement), or
       - before the marker, followed by a blank line (`use X;\n\n<marker>`).
       Then `T` = `short`.
  4. Replace the reported range with `T::class`.
- Multiple fixes in one file are applied in document order; imports added by
  earlier fixes are seen by later ones (a later literal for the same class
  then reuses the import; a later class with the same short name then
  collides and stays fully qualified).
- Results:
  - plain literal, imported via `use Ns\Thing as Other;` → `Other::class`;
  - root-namespace class (`\Thing`, only one backslash) never gets an import
    → `\Thing::class` (unless imported/aliased already);
  - collision with an existing import's local name → `\Full\Name::class`;
  - collision with a class of the same short name in the first namespace →
    `\Full\Name::class`;
  - namespace concat `__NAMESPACE__ . '\Item'` → `Item::class`;
    `__NAMESPACE__ . '\Sub\Item'` → `Sub\Item::class`.

## Options

| Option | Type | Default | Effect |
|---|---|---|---|
| `IMPORT_CLASSES_ON_QF` | bool | true | Fix may add a `use` import for namespaced classes and reference them by short name. |
| `USE_RELATIVE_QF` | bool | false | When importing is otherwise possible, a class below the file's (first) namespace is written as a namespace-relative name instead of being imported. |
| `LOOK_ROOT_NS_UP` | bool | false | Also consider unqualified strings (no backslash), looked up as root-namespace classes (`'Thing'` → `\Thing`). |

Upstream fixtures run with `IMPORT_CLASSES_ON_QF=true, USE_RELATIVE_QF=true,
LOOK_ROOT_NS_UP=true`, except the no-namespace configuration case
(`true, false, false`). Some upstream cases rely on classes declared in a
companion file (project index).

## PHP versions

- Whole rule requires PHP ≥ 5.5. Upstream tests run at the test default
  level (below 7.1, ≥ 5.5).

## Examples

Options: `IMPORT_CLASSES_ON_QF=true`, `USE_RELATIVE_QF=true`,
`LOOK_ROOT_NS_UP=true`.

```php
<?php

namespace Shop\Billing;

use Shop\Billing\Invoice as Bill;
use \ArrayObject as Bag;

class Invoice {}
class Ledger {}

namespace Shop\Billing\Tax;
class Rate {}

namespace Vendor\Kit;
class Ledger {}
class Meter {}

namespace Shop\Billing;

class Clerk {
    public function names() {
        return [
            <weak_warning descr="Use static::class instead.">get_called_class()</weak_warning>,
            get_parent_class(),
            <weak_warning descr="Use \ArrayObject::class instead of the class name string.">'\ArrayObject'</weak_warning>,
            <weak_warning descr="Use \ArrayObject::class instead of the class name string.">'ArrayObject'</weak_warning>,
            <weak_warning descr="Use \Shop\Billing\Invoice::class instead of the class name string.">'Shop\Billing\Invoice'</weak_warning>,
            <weak_warning descr="Use \Shop\Billing\Ledger::class instead of the class name string.">"\\Shop\\Billing\\Ledger"</weak_warning>,
            <weak_warning descr="Use Ledger::class instead of the class name string.">__NAMESPACE__ . '\Ledger'</weak_warning>,
            <weak_warning descr="Use \Shop\Billing\Tax\Rate::class instead of the class name string.">'Shop\Billing\Tax\Rate'</weak_warning>,
            <weak_warning descr="Use \Vendor\Kit\Ledger::class instead of the class name string.">'Vendor\Kit\Ledger'</weak_warning>,
            <weak_warning descr="Use \Vendor\Kit\Meter::class instead of the class name string.">'Vendor\Kit\Meter'</weak_warning>,
            <weak_warning descr="Use \Vendor\Kit\Meter::class instead of the class name string.">'\Vendor\Kit\Meter'</weak_warning>,
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

namespace Shop\Billing;

use Shop\Billing\Invoice as Bill;
use \ArrayObject as Bag;
use Vendor\Kit\Meter;

class Invoice {}
class Ledger {}

namespace Shop\Billing\Tax;
class Rate {}

namespace Vendor\Kit;
class Ledger {}
class Meter {}

namespace Shop\Billing;

class Clerk {
    public function names() {
        return [
            static::class,
            get_parent_class(),
            Bag::class,
            Bag::class,
            Bill::class,
            \Shop\Billing\Ledger::class,
            Ledger::class,
            Tax\Rate::class,
            \Vendor\Kit\Ledger::class,
            Meter::class,
            Meter::class,
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

- `"\\Shop\\Billing\\Ledger"`: the class lives directly in the file's first
  namespace, so the namespace-collision check (`Shop\Billing\Ledger` exists)
  finds the class itself → no import, stays fully qualified.
- `'Shop\Billing\Tax\Rate'`: below the first namespace, no collision →
  relative name `Tax\Rate` (USE_RELATIVE_QF), no import.
- `'Vendor\Kit\Ledger'`: `Shop\Billing\Ledger` exists → stays fully qualified.
- `'Vendor\Kit\Meter'`: imported after the last `use`; the following literal
  for the same class reuses that import.
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
    <weak_warning descr="Use \Shop\Billing\Invoice::class instead of the class name string.">'\Shop\Billing\Invoice'</weak_warning>,
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
    Invoice::class,
    'Local',
];
```

## Divergences

- `__NAMESPACE__ . '\Name'` outside any class-like is resolved against the
  root namespace (the namespace prefix is not applied). Recommendation: use
  the enclosing namespace declaration instead of the enclosing class when
  computing the candidate (equivalent inside classes). Not covered by
  fixtures.
- `__NAMESPACE__ . 'Name'` (no separating backslash) is treated as
  `Ns\Name`, although at runtime it concatenates to `NsName`.
  Recommendation: require the literal to start with a backslash in the
  namespace-concat case.
- Namespace-concat fix with a relative name of 2+ segments separated by 2+
  backslashes (`__NAMESPACE__ . '\A\B\C'`) may add a bogus import
  `use A\B\C;` upstream. Recommendation: in the namespace-concat case never
  import; always produce `<relative name>::class`.
- Double-quoted literals are examined raw: `"Shop\name"` is matched as a class
  name although `\n` is a newline at runtime. Recommendation: skip
  double-quoted literals containing an escape sequence that PHP decodes
  (`\n \t \r \v \e \f \0-7 \x \u`), i.e. any backslash not followed by
  another backslash or by a character without escape meaning.
- The namespace-collision and relative-name checks use the file's *first*
  namespace, not the namespace enclosing the literal. Recommendation: use the
  enclosing namespace (identical for single-namespace files, which is what
  fixtures cover).
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
- **Import matching case (custos diverges):** upstream compares imported
  names and local aliases case-sensitively, so with `use Other\LEDGER;` the
  fix for `'Vendor\Kit\Ledger'` adds `use Vendor\Kit\Ledger;`, a fatal "name
  already in use" error. custos compares both ignoring case (F3 step 1).
- **Leading backslash (custos diverges).** `X::class` never starts with a
  backslash, so rewriting `'\Shop\Jar'` to `\Shop\Jar::class` changes the
  string's value (`'\Shop\Jar' === 'Shop\Jar'` is false): string
  comparisons, array keys and maps built from such literals behave
  differently after the fix. custos still reports these literals but
  without a fix, with the message `Use {name}::class instead of the class
  name string (::class has no leading backslash).` Literals without a
  leading backslash (and `__NAMESPACE__ . '\X'`, whose value has none) keep
  the fix. EA cases affected: `class-name-constant-collisions.php`,
  `configuration-class-reference.php` (listed divergences).
