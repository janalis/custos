---
id: OffsetOperations
group: Probable bugs
kind: semantic
needs: [names, index, hierarchy, types, stubs, flow]
php: { min: "", max: "" }
---

# OffsetOperations

## Summary
Offset access `$x[...]` only works on arrays, strings and objects that
implement offset handling. Using it on a boolean, number or a plain object is
a bug (or a sign of missing type annotations); using an index whose type the
container does not accept (an array as a key, an object as a key) is also
flagged.

Disabled by default.

## Detection
For every array-access expression `C[I]` (also the legacy `C{I}` form and the
push form `C[]`) that is syntactically complete (has its closing bracket and
a container `C`):

### Container types
- **D1** Determine the container type set `S`:
  - First try *value discovery* on `C` (below). If it yields exactly one
    value expression, `S` = inferred type of that value.
  - Otherwise `S` = inferred type of `C` itself.
  - If inference yields nothing, or any part is unknown/unresolvable, `S` is
    empty. Normalise each part: `integer`→`int`, `boolean`/`true`/`false`→
    `bool`, any `T[]` → `array`, `\Closure` (any case) → `callable`, leading
    `\` dropped for built-in names; class names stay fully qualified. Remove
    `self` and `static`.
- **D2** Value discovery of an expression (parentheses stripped):
  - ternary → union of discoveries of its two result branches (for the
    short `?:` form the condition acts as the true branch);
  - `a ?? b` → union of discoveries of `a` and `b`;
  - plain variable inside a function-like: the default value of the
    same-named parameter (if any), plus the right side of every plain `=`
    assignment (not compound, not by-reference) in that function body whose
    left side is textually the same variable (for chained `$a = $b = v` the
    innermost value `v`), each discovered recursively. **Unstable variable** (custos refinement, see Divergences): if the variable is also the operand of `++`/`--` or the target of a compound assignment (`+=`, `.=`, `??=`, …) anywhere in that scope body, the whole result is *unknown*:
    then `S` is empty (no fallback to the inferred type of `C`) → no report;
  - property fetch: the resolved property's default value (unless its text
    ends with the property name), plus plain assignments to the same fetch in
    the current function and in the class constructor;
  - class constant: its value, recursively; global constant (not
    `true`/`false`/`null`): the value of its `define()`;
  - anything else: the expression itself.
  Already-visited expressions are not revisited.
- **D3** Shortcuts on `S` (no report at all):
  - `S` contains `mixed`;
  - `S` is exactly {`string`, `int`} (foreach keys) or exactly {`string`,
    `bool`} (built-ins returning string|false).
- **D4** Then: `callable` is replaced by `array` + `string`; remove `null`,
  `void`, `object` and the empty-set marker. If `S` is now empty → no report.

### Offset support
- **D5** Walk `S`:
  - `array` or `string` → supported; allowed index types gain `string` and
    `int`.
  - any other non-class type (`bool`, `int`, `float`, `resource`, `iterable`,
    …) → **not supported**, stop the walk.
  - a class/interface FQN → for every class/interface found for it in the
    index, check (including inherited members) the methods `offsetGet`,
    `offsetSet`, `__get`, `__set`. For each one present the container is
    supported, and: for `__get`/`__set` allowed index types gain `string` and
    `int`; for `offsetGet`/`offsetSet` they gain the inferred type of that
    method's first parameter (declared type or `@param`; unknown parts
    dropped; nothing when untyped). A class with none of these methods does
    not by itself mark the container as unsupported, but leaves it unsupported
    if nothing else in `S` supports it. A class name not found in the index
    is unresolvable: per D1, `S` is then empty and nothing is reported
    (`$c = new MissingChart(); $c[0]` — the class may simply live under
    another namespace or in an unindexed file).
  Net effect: any scalar in `S` → not supported; otherwise supported iff at
  least one array/string or one class with one of the four methods.
- **D6** Not supported → report the whole access expression (message lists
  `S`); no index check for this node.

### Index type
- **D7** Supported, the allowed index set is non-empty, and an index
  expression `I` exists (not `C[]`): infer `I`'s type, drop unknown parts,
  normalise. If empty → nothing. Remove from it: `mixed`, `null`, every type
  in the allowed set, every class type that extends or implements a class
  type of the allowed set (per the index: a `\Child` index is accepted where
  `\Base` is), and — when the allowed set contains `object` — every class
  type. If the allowed set contains `mixed`, nothing remains. If
  anything remains → report `I`.

## Exceptions (no report)
- **E1** Unknown/partially-unknown container types; `mixed`.
- **E2** `string|int` and `string|false` containers.
- **E3** Containers whose only non-null types are `object`/`null`/`void`.
- **E4** Containers mixing arrays with classes lacking offset methods
  (`\Foo|array`).
- **E5** Index types `null`/`mixed`, or index type unknown.
- **E6** Push form `C[]` on supported containers (no index to check).
- **E7** Objects with an untyped `offsetGet` parameter and no magic
  accessors: allowed set empty → index not checked.
- **E8** Index objects of a subclass (or implementation) of the accepted
  class/interface type.

## Report
- Range:
  - D6: the whole access expression from the start of `C` to the closing
    bracket (`$flag[0]`, `$pdo[]`).
  - D7: the index expression `I` only, without the brackets.
- Severity: error.
- Messages:
  - D6: `'{C}' does not support offset access (types: {S}).` — `{C}` is the
    container source text.
  - D7: `Index of type {types} does not fit the accepted {allowed}.`

## Fix
None.

## Options
None.

## PHP versions
The legacy `$s{0}` syntax is accepted and treated like `$s[0]` (it appears in
the upstream fixture, which runs at the test default level < 7.1; it was
removed in PHP 8.0).

## Examples

```php
<?php
class Registry {
    /** @param string $name */
    public function offsetGet($name) { return null; }
}
class Dyn {
    public function __set($k, $v) {}
}

function demo(array $rows, int $limit) {
    $title = 'abc';
    echo $title[1];
    $title[<error descr="Index of type array does not fit the accepted string|int.">array_keys($rows)</error>] = 'z';
    $title[] = 'x';

    $count = 10;
    <error descr="'$count' does not support offset access (types: int).">$count[2]</error> = 1;

    $reg = new Registry();
    echo $reg['db'];
    echo $reg[<error descr="Index of type int does not fit the accepted string.">7</error>];

    $dyn = new Dyn();
    $dyn[<error descr="Index of type \ArrayObject does not fit the accepted string|int.">new ArrayObject()</error>] = 1;

    $plain = new DateTime();
    <error descr="'$plain' does not support offset access (types: \DateTime).">$plain[]</error> = 1;

    foreach ($rows as $k => $v) {
        echo $k[0];
    }
    $tail = substr('abc', 1);
    echo $tail[0];
}

/**
 * @param mixed $m
 * @param \Countable|array $c
 * @param null|array $n
 */
function ok($m, $c, $n) {
    return [$m['a'], $c['a'], $n['a']];
}
```

## Divergences
- **Unstable variables — custos refinement, not upstream.** Value discovery ignores `++`/`--` and compound assignments upstream, so a variable later incremented or extended is analysed with its initial value only. custos makes the result unknown (`S` empty, no report), as in the shared value discovery of `CallableMethodValidity`. No upstream fixture relies on such a variable; recorded in `docs/decisions.md` ("Spec-level false positives").
- Built-in return types are stub-dependent. Upstream's fixture expects no
  report for `explode(...)[0]`-style containers whose stub type is
  `string[]|false` (normalised {`array`, `bool`}), which the algorithm above
  would report. Recommendation: in D3 also treat exactly {`array`, `bool`}
  as a shortcut (no report) — or have stubs give `explode` a plain `array`
  return type — so the conformance case passes.
- Index arithmetic: upstream infers `$key + 1` (with `$key` an array key) as
  an acceptable index type; make sure `int + int` infers `int` (not
  `int|float`), or treat arithmetic results as unknown, to avoid false
  positives.
- Subtype indexes (custos diverges from upstream). Upstream checks index
  types by plain set membership, so with `offsetGet(Key $k)` an index of a
  subclass `SpecialKey` is reported although PHP accepts it. custos accepts
  subtypes of the allowed class types (D7).
- **Complete fallback, failure markers, bool keys (custos diverges).**
  Found on PrestaShop, Matomo and phpMyAdmin (736 reports, about 410 of
  them false). (1) The single-value fallback of D1 now needs *every* value
  of the container (`PossibleValuesComplete`): a parameter default
  (`function copy($tables = false)` called with arrays) or `$last = false`
  reassigned from a `foreach` value is only one of the values, so it no
  longer types the container as `bool`. (2) In D4 a `bool` next to `array`,
  `string` or `callable` is a failure marker (`array|false` from a lookup,
  `Db::getRow(): array|bool|object|null`) and is dropped like `null`.
  (3) In D7 a `bool` index is accepted where `int` is: PHP casts `true`/
  `false` keys to 1/0. A `float` index is accepted there too: PHP truncates it
  to int (8.1 deprecates fractional values; pChart's `$Units[floor($v)]`). A variable read before its first assignment is no
  longer typed by that later assignment.
- **Native offset classes (custos diverges).** `DOMNodeList`,
  `DOMNamedNodeMap`, `ResourceBundle`, `Dom\NodeList`, `Dom\NamedNodeMap`,
  `Dom\HTMLCollection`, `FFI\CData` and their subclasses support `$o[…]`
  natively although the stubs declare no `offsetGet()`: supported, with
  `string|int` keys (Magento locale lists, Joomla DOM templates).
- **Loose PHPDoc unions (custos diverges).** When `S` contains `array` or
  `string` next to other scalar members (`@return
  array|int|string|float|bool`, 3,416 Magento test findings from one
  helper) and the variable's type without PHPDoc is unknown, nothing is
  reported: the documentation admits offset access and the other members
  are not confirmed by native types. Native unions (`int|array $v`) and
  documented types without an array/string member (`@param \stdClass`)
  are still reported.
- **Round-4 refinements (custos diverges).** Found on API Platform,
  Shopware, Mautic, Akeneo and Kimai:
  - `iterable` is `array` plus `\Traversable` (D4), not an unsupported
    scalar: `private iterable $handlers = []; $this->handlers[$k] = $h;`.
  - A class member that is an interface or a non-final abstract class
    without `offsetGet`/`offsetSet`/`__get`/`__set` empties S (like an
    unresolvable class): its implementations may be `ArrayAccess`
    (Laravel's `Application $app; $app['config']`).
  - D6 is not reported for reads inside `isset()`, `empty()` or on the
    left of `??` (also as the base of a longer chain) when S has an array
    or string member: PHP yields null there without an error, and the union
    is a deliberate lookup (`array|int $counts; $counts[$id] ?? 0`).
  - D7: an index type containing `mixed` is not checked (any key may
    fit), and `void` (`@return string|void`) is dropped like `null`.
- **Round-5 refinements (custos diverges).** A boolean next to a class is a
  failure marker too (`simplexml_load_string()` returns
  `SimpleXMLElement|false`; ten Moodle reports). D7 skips an index whose
  documented union admits an accepted type (`@return string[]|string`,
  depending on the input) unless native declarations type the index.
- **Unresolvable ancestors (custos diverges).** A class whose parent,
  interface or trait (at any depth) does not resolve is treated like an
  unresolvable class (D1): the missing ancestor may implement
  `ArrayAccess` (Webklex's `FolderCollection` extends Laravel's
  `Collection`, not indexed in Dolibarr's bundled copy).
