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
- **D2 (inherited duplicate)** Otherwise, when the property is not static
  (custos, see Divergences), `D` and `OD` both exist,
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
  Also nothing when the property has a declared type (custos, see
  Divergences). Also nothing when the assigned value is exactly a variable
  naming a non-variadic parameter of this constructor whose default value
  is equivalent to `D` (`private $tags = [];` with
  `__construct(array $tags = [])` and `$this->tags = $tags;`): the property
  default is the same "not provided" value as the parameter's, and it is the
  one an object created without running the constructor keeps (custos
  diverges, see Divergences). Otherwise, when `REPORT_DEFAULTS_FLAWS` is on, report
  pattern **O** on that property's default `D`.
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
- **E6** Overrides by a constructor parameter whose default equals the
  property default (D7), e.g. `private $strict = false;` with
  `__construct(bool $strict = false) { $this->strict = $strict; }`. A
  parameter without default, with a different default, or an expression
  built from the parameter (`(bool) $strict`, `$strict ?? false`) does not
  qualify.

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
  - S: `Default repeats the inherited value; drop the re-declaration.`
  - O: `Default is always replaced by the constructor; remove it.`
  - W: `Assignment writes the property's default value; remove it.`

## Fix

Only patterns N and O have a fix; S and W have none.

- **F1** Delete everything from immediately after the property's name
  (`$name`) through the end of the default value — the whitespace, any
  comments, the `=` and the value. The terminator (`,` or `;`) stays.
  - `protected $cache = null;` → `protected $cache;`

  - ```text
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
    protected $items = <weak_warning descr="Default repeats the inherited value; drop the re-declaration.">[]</weak_warning>;
    private $secret = 'x';
    public static $currency = 'EUR';
    protected $tagClass = <weak_warning descr="Default repeats the inherited value; drop the re-declaration.">Tag::class</weak_warning>;
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

Defaults mirrored by a constructor parameter's default (E6):

```php
<?php
final class Job
{
    /** Kept for payloads serialized before this field existed. */
    private $context = [];
    private $retry = false;
    private $queue = 'default';
    private $level = <weak_warning descr="Default is always replaced by the constructor; remove it.">1</weak_warning>;

    public function __construct(array $context = [], bool $retry = false, string $queue = 'low', int $level = 1)
    {
        $this->context = $context;
        $this->retry = $retry;
        $this->queue = $queue;
        $this->level = $level * 2;
    }
}
```

```php
<?php
final class Job
{
    /** Kept for payloads serialized before this field existed. */
    private $context = [];
    private $retry = false;
    private $queue;
    private $level;

    public function __construct(array $context = [], bool $retry = false, string $queue = 'low', int $level = 1)
    {
        $this->context = $context;
        $this->retry = $retry;
        $this->queue = $queue;
        $this->level = $level * 2;
    }
}
```

## Divergences

- **Default mirrored by the parameter default (custos diverges).**
  Upstream reports (and its fix removes) `private $payload = [];` when the
  constructor runs `$this->payload = $payload;` with `array $payload = []`.
  Objects are also created without the constructor (`unserialize()` of a
  payload written before the property existed, which then stays `null`;
  `ReflectionClass::newInstanceWithoutConstructor()`; ORM hydration and
  proxies), and those see the class default. When the class default and
  the parameter default are the same value, the author has stated one
  "absent" value for both paths; removing it turned an `array` getter into
  a `TypeError` on old messages (found on real code). custos keeps such
  defaults (E6). It does not go further: a default the constructor
  replaces with an unrelated value (`$this->queue = $queue` with a
  different parameter default, `$this->level = $level * 2`) is still
  reported, since nothing in the code ties the default to the bypass paths,
  and typed properties are exempt anyway (see below).

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
- **Typed properties keep their default (custos diverges).** Upstream
  reports (and its fix removes) the default of a typed private property the
  constructor overwrites (`private array $items = [];`). Without a default
  a typed property is *uninitialised*, so objects created without the
  constructor (unserialize of older payloads, `newInstanceWithoutConstructor`,
  ORM hydration, proxies) throw "must not be accessed before
  initialization" where they used to see `[]`. custos skips pattern O for
  typed properties.
- **Pattern S message (custos).** Removing only the default of a
  re-declared property does not inherit the parent's default: the property
  becomes null (untyped) or uninitialised (typed). The message therefore
  advises dropping the whole re-declaration.
- **Static re-declarations (custos diverges).** Upstream reports a static
  property re-declared with the parent's default. Re-declaring a static
  property gives the subclass its own storage (a per-class registry or
  cache); dropping it makes the subclass share — and overwrite — the
  parent's value. custos skips static properties in D2.
- **`self::`/`parent::` in defaults (custos diverges).** The class-reference
  guard of D2 treated `self`/`parent` as unresolvable, so `protected $kind =
  self::KIND;` re-declared in a child that overrides `KIND` was reported as
  repeating the inherited value although it names the child's constant.
  custos resolves `self` to the class declaring the default and `parent` to
  that class's parent, so such re-declarations are not reported (they
  denote different classes in child and parent).
- **Calls before the write (custos diverges).** Pattern O is skipped when
  a statement before the constructor assignment, or its right-hand side,
  can run the object's own code: a method call on `$this`, a `parent::`,
  `self::` or `static::` call, or `$this` passed to a call. That code may
  read the default (Composer's `Pool::__construct()` calls
  `$this->setPackages()` before assigning the other properties), so
  removing it could change behaviour.
