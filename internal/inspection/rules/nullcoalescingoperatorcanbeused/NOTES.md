# NullCoalescingOperatorCanBeUsed implementation notes

## Type inference for G3

G3 needs the inferred type of the candidate property. It is taken from:

- declared property types;
- `@var` doc comments on properties (`null|Foo`, `string`, …);
- inline `@var` doc comments naming a variable's class
  (`/** @var Foo $x */`);
- element types of array-typed variables (`Foo[]` gives `Foo`);
- chained accesses: for `$x->a->b` where `a` is `null|Foo` and `Foo::$b` is
  `string`, the result is `string` (the intermediate's nullability does not
  carry over to the final type).

An unresolved base class or property yields no known type, so G3 does not
apply.
