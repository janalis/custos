---
id: LongInheritanceChain
group: Architecture
kind: semantic
needs: [names, hierarchy, index, stubs]
php: { min: "", max: "" }
---

# LongInheritanceChain

## Summary

Deep `extends` chains spread behaviour over many levels and make a class hard
to understand. The rule counts the ancestors of a class and complains when the
chain reaches a threshold, with carve-outs for exception hierarchies, test
cases and well-known framework base classes.

## Detection

Applies to every named class-like declaration that can have a parent class
(in practice: classes, including abstract ones).

- **D1** Skip when the class's short name ends with `Exception`
  (case-sensitive suffix), or the class is in a test context (T below).
- **D2** Let `P1` be the resolved direct parent class. If `P1` exists, the
  class is **not** abstract and `P1` **is** abstract → no report.
- **D3** Walk up the chain, counting:

  ```text
  count = 0; cur = P1
  while cur exists and cur is not the class itself:
      next = resolved parent of cur
      count = count + 1
      if next exists:
          if FQN(next) is a stop class (S below): count = count + 1; stop walking
          if short name of next ends with "Exception": no report at all (abort)
      cur = next
  ```

  Notes that matter:
  - the direct parent `P1` itself is never tested against S or the
    `Exception` suffix — only ancestors from the grandparent upwards are;
  - when a stop class is met it is counted, but its own ancestors are not;
  - an unresolvable parent simply ends the walk (it is not counted).
  So `count` is the number of resolvable ancestors, truncated at the first
  stop class at grandparent level or above.
- **D4** Report when `count >= COMPLAIN_THRESHOLD` and the class is not
  marked `@deprecated` in its doc comment.

**S — stop classes** (exact FQNs): `\PHPUnit_Framework_TestCase`,
`\PHPUnit\Framework\TestCase`, `\yii\base\Component`, `\yii\base\Behavior`,
`\CComponent`, `\Zend\Form\Form`, `\Phalcon\Di\Injectable`.

**T — test context**: the file path ends with `Test.php`, `Spec.php` or
`.phpt`, or contains `/Fixtures/`; or the class FQN ends with `Test`, or
contains `\Tests\` or `\Test\`.

## Exceptions (no report)

- **E1** Classes named `…Exception`, and classes having an ancestor at
  grandparent level or above named `…Exception`.
- **E2** Concrete classes directly extending an abstract class.
- **E3** `@deprecated` classes.
- **E4** Test-context classes.
- **E5** Chains shorter than the threshold.
- **E6** A class that (directly or via a cycle) extends itself: the walk stops
  when it comes back to the class.

## Report

- Range: the class name identifier.
- Severity: **info** (weak warning). Upstream registers this problem with a
  weak-warning level even though the catalogue default is `warning`; the
  fixtures expect weak warnings.
- Message: `{count} levels of parent classes; prefer composition over deep
  inheritance.`

## Fix

None.

## Options

| Option | Type | Default | Effect |
|---|---|---|---|
| `COMPLAIN_THRESHOLD` | int | 3 | Minimum ancestor count (inclusive) that triggers a report. |

## PHP versions

No gating.

## Examples

Default threshold 3:

```php
<?php
namespace App {
    class Node {}
    class Branch extends Node {}
    class Twig extends Branch {}
    class <weak_warning descr="3 levels of parent classes; prefer composition over deep inheritance.">Leaf</weak_warning> extends Twig {}

    /** @deprecated */
    class OldLeaf extends Twig {}
    class TwigException extends Twig {}

    abstract class Base extends Branch {}
    class Concrete extends Base {}

    class Err1 extends \RuntimeException {}
    class Err2Exception extends Err1 {}
    class Err3 extends Err2Exception {}
    class Err4 extends Err3 {}
}

namespace yii\base {
    class Component {}
}
namespace App\Widgets {
    class Panel extends \yii\base\Component {}
    class FancyPanel extends Panel {}
    class <weak_warning descr="3 levels of parent classes; prefer composition over deep inheritance.">MegaPanel</weak_warning> extends FancyPanel {}
}
```

(`Err4`: walking from `Err3`, the grandparent `Err2Exception` ends with
`Exception` → aborted. `Concrete`: concrete class over an abstract parent.)

## Divergences

- Cycles not passing through the class itself (`A extends B`, `B extends C`,
  `C extends B`) would loop forever upstream. Recommendation: stop the walk
  on any already-visited class.
