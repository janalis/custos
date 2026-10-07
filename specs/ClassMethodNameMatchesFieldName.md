---
id: ClassMethodNameMatchesFieldName
group: Confusing constructs
kind: semantic
needs: [names, hierarchy, index, types]
php: { min: "", max: "" }
---

# ClassMethodNameMatchesFieldName

## Summary
When a method and a property share a name, `$obj->name()` and
`($obj->name)()` mean different things. This is especially confusing when the
property holds a callable, and when the property's type is unknown the reader
cannot tell which one is meant.

## Detection
- **D1** A method declared in a class, abstract class, trait or enum (any
  visibility, static or not). Methods of interfaces are skipped.
- **D2** Look up a **property** (not a class constant) whose name equals the
  method name exactly (case-sensitive, without `$`) in the containing class,
  including properties inherited from ancestor classes and properties of used
  traits (any visibility, static or not). No match → nothing.
- **D3** Compute the property's type set: union of
  - its declared type (typed property),
  - the types of an `@var` tag in the comment immediately preceding the
    declaration — both doc comments `/** … */` **and** plain block comments
    `/* … */` are honoured,
  - the type of its default value, if any (e.g. `= 0` → int, `= null` → null,
    `= []` → array).
  Drop unknown/unresolvable entries.
- **D4** If the set is empty → report "type unknown" (D4 message).
- **D5** Otherwise, if any type in the set is `callable` or the class
  `\Closure` (compare case-insensitively; a bare `closure`/`Closure` in a
  docblock of a file in the global namespace resolves to `\Closure`) → report
  "callable" (D5 message). Any other known type → no report.

## Exceptions (no report)
- **E1** Interface methods.
- **E2** A same-named class constant (`const name`) is not a property.
- **E3** A same-named property with a known, non-callable type
  (e.g. `@var int`, `= 'x'`, `?string`).

## Report
- Range: the method name identifier.
- Severity: info (both cases).
- Messages:
  - D4: `A property with this name exists and its type is unknown; rename
    the method or type the property.`
  - D5: `A callable property with this name exists; rename the method (for
    example with a get/is/has prefix).`

## Fix
None.

## Options
None.

## PHP versions
No gating.

## Examples

```php
<?php
class Account {
    private $owner;
    public function <weak_warning descr="A property with this name exists and its type is unknown; rename the method or type the property.">owner</weak_warning>() {}
}

class Hooks {
    /* @var callable */
    protected $onSave;
    /** @var \Closure|null */
    protected $onLoad;
    public function <weak_warning descr="A callable property with this name exists; rename the method (for example with a get/is/has prefix).">onSave</weak_warning>() {}
    public static function <weak_warning descr="A callable property with this name exists; rename the method (for example with a get/is/has prefix).">onLoad</weak_warning>() {}
}

class Child extends Account {
    public function <weak_warning descr="A property with this name exists and its type is unknown; rename the method or type the property.">owner</weak_warning>() {}
}

class Counter {
    private $total = 0;
    private ?string $label;
    const size = 3;
    public function total() {}
    public function label() {}
    public function size() {}
}

interface Named {
    public function name();
}
```

## Divergences
- Upstream takes the property type from the IDE's type engine; whether types
  inferred from `$this->prop = …` assignments inside the class also count is
  not observable in fixtures. Recommendation: use only D3 sources (declared
  type, preceding `@var` comment, default value).
