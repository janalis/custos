---
id: PropertyInitializationFlaws
group: Unused
kind: semantic
needs: [names, hierarchy, index]
php: { min: "", max: "" }
---

# PropertyInitializationFlaws

## Summary
Property defaults and constructor assignments that do nothing: an explicit
`= null` default (untyped properties are already `null`), a default that
merely repeats the inherited one, a default on a private property that the
constructor unconditionally overwrites, and a constructor assignment that
writes exactly the default value.

## Detection
Two independent checks.

### Check 1 — property defaults (only when `REPORT_DEFAULTS_FLAWS` is on)
Visit every property declaration (each variable of a `public/protected/
private [static] [type] $a = …, $b = …;` declaration separately; class
constants are not properties). Let `D` be its default value expression
(absent if none). Let `Par` be the **direct** parent class of the containing
class (resolved; none for traits, unextended classes, or unresolvable
parents) and `O` the property with the same name found in `Par` or anything
`Par` inherits (its own ancestors and used traits); `OD` its default.

- **D1 (null default)** `D` is the constant `null` (case-insensitive, an
  optional leading `\` allowed) → report pattern **N** on `D`, unless E1
  applies. (Independent of any parent.)
- **D2 (inherited duplicate)** Otherwise, when `D` and `OD` both exist,
  `O` is **not** `private`, and `D` is *equivalent* to `OD` (same node kind
  and structurally identical ignoring whitespace/comments, or identical
  source text) → candidate.
  - Class-reference guard: collect the fully-qualified names that every class
    reference inside `OD` resolves to (e.g. the `Foo` in `Foo::class`,
    `Foo::BAR`; unresolvable references count as one "unknown" entry). If that
    set is non-empty, also collect the same for `D`; report only when every
    name from `D` is already in `OD`'s set (identical text resolving to
    different classes in different namespaces → no report).
  - Report pattern **S** on `D`.

### Check 2 — constructor (only when `REPORT_INIT_FLAWS` is on)
Visit every method named `__construct` (any letter case, as PHP compares
method names) whose
containing type is a class (not an interface or trait) and whose body has at
least one statement.

- **D3** Build the candidate map from the class's **own** declared
  properties (not inherited, not from traits) that are `private` and not
  `static`: name → default `D`, where a missing default and a `null` default
  are both recorded as "no default". If the map is empty, stop.
- **D4** For each statement that is a **direct** child of the constructor
  body and is an expression statement whose expression is a plain `=`
  assignment (by-reference `= &` included; compound `.=` etc. excluded;
  for a chained assignment only the outermost target counts):
  - the target is a property fetch whose object is literally `$this`
    (`$this->name`, with a static identifier name); `self::$x`,
    `$this->a->b`, `$other->x`, dynamic names → skip;
  - the property name is in the map.
  Assignments nested inside `if`, loops, `try`, closures, etc. are not
  considered.
- **D5 (writes the default)** If (the map has "no default" for it and the
  assigned value is the constant `null`) or (it has default `D` and the
  assigned value is equivalent to `D`) → report pattern **W** on the
  statement, unless E1 applies to that property. Continue with the next
  statement.
- **D6** Else, if the map has "no default" → nothing.
- **D7 (overridden default)** Else, if the assigned value contains (at any
  depth) a property fetch equivalent to the target (`$this->name`) → nothing
  (the default is used to compute the new value). Also nothing when an
  earlier direct statement of the constructor body contains a `return`
  (nested functions/closures excluded): the constructor may leave before the
  assignment and keep the default (custos diverges, see Divergences).
  Otherwise, when `REPORT_DEFAULTS_FLAWS` is on, report pattern **O** on
  that property's default `D`.
- **Name case.** Wherever this rule compares two expressions for
  equivalence, the names PHP resolves case-insensitively — function and
  method names, class names in calls, `new`, `instanceof` and `::`
  accesses, and keywords — compare case-insensitively (`Cache::$map['k']`
  matches `cache::$map['k']`, `$o->Name()` matches `$o->name()`); variable,
  property and constant names stay case-sensitive.

## Exceptions (no report)
- **E1** Nullable typed properties: when the language level is **7.4 or
  higher** and the property's declared type contains `null` (`?T`,
  `T|null`, `null`) or is/contains `mixed`, patterns N and W are not
  reported for it. Below 7.4 typed properties get no special treatment.
- **E2** D2 when the parent's property is `private`.
- **E3** Non-private or static properties in Check 2 (constructor may be
  bypassed / static state).
- **E4** Assignments not directly in the constructor body (D4), assignments
  in other methods.
- **E5** Overrides that reuse the property (`$this->p = array_merge($this->p, …)`).

## Report
- Ranges:
  - N: the `null` default token.
  - S: the default value expression `D`.
  - O: the default value expression `D` of the property (in the
    declaration, not in the constructor).
  - W: the whole assignment statement in the constructor, from `$this`
    through the terminating `;` inclusive.
- Severity: info for all (fixture markup `weak_warning`).
- Messages:
  - N: `Explicit null default is redundant; remove it.`
  - S: `Default repeats the inherited value; remove it.`
  - O: `Default is always replaced by the constructor; remove it.`
  - W: `Assignment writes the property's default value; remove it.`

## Fix
Only patterns N and O have a fix; S and W have none.

- **F1** Delete everything from immediately after the property's name
  (`$name`) through the end of the default value — the whitespace, any
  comments, the `=` and the value. The terminator (`,` or `;`) stays.
  - `protected $cache = null;` → `protected $cache;`
  - ```
    private $mode // legacy
        = null;
    ```
    → `private $mode;` (the comment between name and `=` is removed too).
  - `private $items = [];` (overwritten in constructor) → `private $items;`

## Options
| Option | Type | Default | Effect |
|---|---|---|---|
| `REPORT_DEFAULTS_FLAWS` | bool | `true` | Enables Check 1 (N, S) and pattern O of Check 2. |
| `REPORT_INIT_FLAWS` | bool | `true` | Enables Check 2 (W, and O when the other option is on). |

## PHP versions
- E1 only applies at language level ≥ 7.4. Upstream cases without an explicit
  level run at the IDE test default (5.6–7.0), the typed-property case runs
  at 7.4. Typed properties are parsed at any level.

## Examples

```php
<?php
namespace Shop;

class Tag {}

class BaseCart
{
    protected $items = [];
    private $secret = 'x';
    public static $currency = 'EUR';
    protected $tagClass = Tag::class;
    protected $notes = <weak_warning descr="Explicit null default is redundant; remove it.">NULL</weak_warning>;
}

class Cart extends BaseCart
{
    protected $items = <weak_warning descr="Default repeats the inherited value; remove it.">[]</weak_warning>;
    private $secret = 'x';
    public static $currency = <weak_warning descr="Default repeats the inherited value; remove it.">'EUR'</weak_warning>;
    protected $tagClass = <weak_warning descr="Default repeats the inherited value; remove it.">Tag::class</weak_warning>;
    protected $notes = 'n/a';

    private $total = <weak_warning descr="Default is always replaced by the constructor; remove it.">0</weak_warning>;
    private $lines = [];
    private $count = 0;
    private $owner;
    private $history = [];
    private $later = 'a';
    protected $shown = 1;

    public function __construct($owner)
    {
        $this->total = 100;
        <weak_warning descr="Assignment writes the property's default value; remove it.">$this->lines = [];</weak_warning>
        <weak_warning descr="Assignment writes the property's default value; remove it.">$this->owner = null;</weak_warning>
        $this->count = $this->count + 1;
        $this->history = array_merge($this->history, [$owner]);
        if ($owner) {
            $this->later = 'b';
        }
        $this->shown = 2;
    }

    public function reset()
    {
        $this->lines = [];
    }
}
```

```php
<?php
namespace Shop;

class Tag {}

class BaseCart
{
    protected $items = [];
    private $secret = 'x';
    public static $currency = 'EUR';
    protected $tagClass = Tag::class;
    protected $notes;
}

class Cart extends BaseCart
{
    protected $items = [];
    private $secret = 'x';
    public static $currency = 'EUR';
    protected $tagClass = Tag::class;
    protected $notes = 'n/a';

    private $total;
    private $lines = [];
    private $count = 0;
    private $owner;
    private $history = [];
    private $later = 'a';
    protected $shown = 1;

    public function __construct($owner)
    {
        $this->total = 100;
        $this->lines = [];
        $this->owner = null;
        $this->count = $this->count + 1;
        $this->history = array_merge($this->history, [$owner]);
        if ($owner) {
            $this->later = 'b';
        }
        $this->shown = 2;
    }

    public function reset()
    {
        $this->lines = [];
    }
}
```

Class-reference guard and typed properties (PHP 7.4):

```php
<?php
namespace A { class Widget {} class Base { protected $kind = Widget::class; } }
namespace B {
    class Widget {}
    class Child extends \A\Base { protected $kind = Widget::class; }
    class Typed
    {
        private ?Widget $w = null;
        private mixed $any = null;
        private ?int $n;
        public function __construct($seed = 0) { $this->n = null; }
    }
}
```
No findings.

## Divergences
- **Early return before the assignment (custos diverges):** upstream
  reports a default as always replaced even when the constructor can
  `return` before assigning it (`if (!$data) { return; } $this->data =
  $data;`), where the default is the value kept on that path; removing it
  leaves the property uninitialised (an error for a typed property). custos
  skips pattern **O** after any earlier statement containing a `return`.
- A property assigned several times directly in the constructor yields one
  O finding per assignment on the same default upstream. Recommendation:
  de-duplicate by range.
- Constructor-promoted parameters with a `null` default are not property
  declarations for this rule; recommendation: do not report them (no
  fixture).
- Pattern W for a non-nullable typed property without default
  (`private int $n; … $this->n = null;`) is reported upstream (it is a type
  error anyway). Keep.
- **Constructor name case — custos diverges from upstream.** Upstream only
  visits a constructor spelled exactly `__construct`; PHP treats
  `__Construct` the same, so custos matches the name case-insensitively.
- **Name case (custos diverges).** Upstream compares the expressions
  textually, so operands that differ only in the case of a function, method
  or class name (`Stats::$n` vs `stats::$n`), which PHP treats as the same,
  are not recognised as equivalent. custos folds the case of those names
  (Detection, "Name case").
