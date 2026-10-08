---
id: MissingIssetImplementation
group: Probable bugs
kind: semantic
needs: [names, index, hierarchy, types]
php: { min: "", max: "" }
---

# MissingIssetImplementation

## Summary
`isset($obj->name)` / `empty($obj->name)` on a property the class does not
declare can only answer correctly if the class implements `__isset()`. When it
does not (typically classes with `__get` only), the check is always
false/empty, which is a silent logic bug.

## Detection
For every argument `A` of `isset(...)` and of `empty(...)`:
- **D1** `A` is directly an instance property fetch `$base->name` using the
  plain `->` operator (nullsafe `?->` is not considered) with a **static
  identifier** name (not `$o->$n`, not `$o->{$expr}`).
- **D2** The property does not resolve: no property declaration (own,
  inherited, from traits, or a documented `@property*` tag) is found for it
  by normal member resolution.
- **D3** `$base` is not literally the text `$this`.
- **D4** Infer the type of `$base`; drop unknown parts and normalise each to
  a fully-qualified name. For each class-like type (starting with `\`) that is
  not `\SimpleXMLElement`, `\stdClass` or `\DOMDocument`:
  - look the class up in the project index (FQN lookup, case-insensitive);
    if not found → skip this type;
  - if the found class' own FQN is one of the three names above → skip;
  - if the class (including inherited members) declares no property `name`
    **and** has no `__isset` method (own or inherited, including from traits)
    → report `A` and stop processing the remaining types of `A`.

## Exceptions (no report)
- **E1** Properties declared on the class or its ancestors (any visibility).
- **E2** Classes having `__isset` somewhere in their hierarchy.
- **E3** Dynamic property names (`$o->{$k}`, `$o->$k`), static properties
  (`A::$p`, `A::${$p}`), nullsafe fetches.
- **E4** `$this->prop`.
- **E5** `stdClass`, `SimpleXMLElement`, `DOMDocument` (any letter case,
  e.g. `new \StdClass()`).
- **E6** Unresolvable types/classes.
- **E7** Property fetches nested deeper inside an argument
  (`isset($a[$o->x])`) — only direct arguments are checked.

## Report
- Range: the whole argument `A` (`$base->name`).
- Severity: error.
- Message: `{Type} has no __isset(); this isset/empty check is always false.`
  where `{Type}` is the normalised FQN as inferred (e.g. `\Acme\Bag`).

## Fix
None.

## Options
None.

## PHP versions
None.

## Examples

```php
<?php
class Profile
{
    public $nick;
    private $secret;
}

class Lazy
{
    public function __get($k) { return 1; }
    public function __isset($k) { return true; }
}

class Bag
{
    public function __get($k) { return null; }

    public function probe(): array
    {
        $p = new Profile();
        $l = new Lazy();
        $b = new Bag();
        $o = new \STDCLASS();
        $k = 'x';
        return [
            isset($p->nick),
            isset($l->anything),
            isset(<error descr="\Bag has no __isset(); this isset/empty check is always false.">$b->color</error>, $b->{$k}),
            empty(<error descr="\Bag has no __isset(); this isset/empty check is always false.">$b->size</error>),
            empty($b->$k),
            isset($this->dynamic),
            isset($o->field),
            isset(Bag::${$k}),
        ];
    }
}
```

## Divergences
None known.
- **Undecidable receivers (custos diverges).** Upstream reports as soon as
  one class of the receiver's type lacks `__isset()`, with severity error
  and "always false". That is wrong when another possible value can carry
  the property, so custos reports only when *every* non-null member of the
  type is a resolvable, concrete class (not an interface, abstract class,
  trait or enum) without the property, without `__isset()` and without
  `#[\AllowDynamicProperties]` in its hierarchy (attributes are seen on
  declarations in the analysed file; the index does not record them for
  other files). A union with `object`, `mixed`, an array or scalar type,
  `stdClass`/`SimpleXMLElement`/`DOMDocument`, or an unresolvable class
  yields no report. Interfaces and abstract classes are skipped because
  their implementations may declare `__isset()`.
- **Dynamic properties (custos diverges).** Without `__set()`, a write
  `$o->name = …` creates a real (dynamic) property that `isset()` does see.
  When the class has no `__set()` and the analysed file writes a property
  of that name on a receiver other than `$this`, or a property with a
  computed name (`$entity->$field = $value`, an importer), the check is not
  reported: it is not always false (PrestaShop: 63 → 1).
- **Dynamic and unresolvable ancestors (custos diverges).** Subclasses of
  `stdClass` (always dynamic, Joomla `Table`) and of `SimpleXMLElement`
  (native isset handler) are skipped like those classes themselves; so is
  a class whose hierarchy does not fully resolve (a missing parent may
  declare the property or `__isset()`).
- **Subclass instances (custos diverges from D4).** A value typed as a
  non-final class may be an instance of a subclass that declares the
  property, `__isset()` or `#[\AllowDynamicProperties]`
  (`isset($fault->errorcode)` on `Exception` where Moodle's exceptions
  declare it, `isset($node->tagName)` on `DOMNode` — `DOMElement` has it);
  such checks are not reported. Descendants come from the project index and
  the stubs.
