---
id: ArgumentCountMismatch
group: Probable bugs
kind: semantic
needs: [names, index, hierarchy, types, stubs]
php: { min: "", max: "" }
---

# ArgumentCountMismatch

## Summary

Report a call that leaves a required parameter without an argument, and a
call to a PHP function or method that passes more arguments than it
declares. Since PHP 7.1 a missing argument throws an `ArgumentCountError`
for functions of the project, and PHP's own functions reject both missing
and extra arguments.

## Detection

- D1. A function call, method call, static call or object construction
  whose callee resolves, as listed in D3, to known declarations.
- D2. Arguments bind as PHP binds them: positional arguments fill the
  parameters in order, named arguments fill the parameter of that name. A
  `new` written without parentheses passes no argument.
- D3. The declarations a call may bind to:
  - a function call: the single declaration of the function the call
    reaches (the namespaced function, else the global fallback);
  - a construction: the constructor the class runs, inherited or its own;
  - an instance method call: the method as each class of the receiver's
    type resolves it, plus the method as every indexed subtype of those
    classes resolves it, since the object may be any of them;
  - a `static::` call: the method of the enclosing class and of its
    indexed subtypes;
  - a call through a class name, `self` or `parent`: the method that class
    resolves.
  Subtypes are not consulted when the method is private or final, or its
  class is final.
- D4. Too few: every declaration of D3 leaves a parameter that has no
  default value and is not variadic without an argument. The message names
  the parameters the nominal declaration leaves unbound.
- D5. Too many: every declaration of D3 is declared by PHP itself (a
  builtin function, or a method or constructor of a builtin class), none
  has a variadic parameter, and the call passes more positional arguments
  than the declaration with the most parameters accepts.
- D6. A builtin's parameters are those it declares at the configured PHP
  version.

## Exceptions (no report)

- E1. A call that unpacks arguments (`...$args`), a first-class callable
  (`f(...)`), or a named argument that matches no parameter or is collected
  by a variadic parameter.
- E2. Callees that cannot be resolved: a dynamic name, a receiver of unknown
  or mixed type, a union receiver whose classes disagree on the parameter
  list, a method not found (served by `__call` or `__callStatic`), a method
  declared only by a `@method` tag, a method the caller cannot see while the
  class declares `__call` or `__callStatic`, a class whose ancestors are not
  all indexed, a trait method called through `$this`, `self` or `static`.
- E3. Uncertain declarations: a function the project declares more than
  once (conditionally loaded variants), a project function that shares the
  name of a function PHP declares in some version (a polyfill), a class
  declared more than once, a constructor in an abstract class, interface or
  enum.
- E4. An override in a subtype that accepts the call (fewer required
  parameters, a variadic parameter) makes the call valid.
- E5. Extra arguments to a function or method of the project: PHP allows
  them, and the callee may read them with `func_get_args()`.
- E6. Builtins below PHP 8.0, whose parameter counts are not declared in a
  form the stubs describe exactly; builtins of extensions not bundled with
  PHP; and builtins whose stub signature is known to differ from the one
  PHP enforces (overloaded forms written as one signature, such as
  `array_keys`, `strtr`, `implode`, `mt_rand` or `stream_filter_append`).

## Report

- Range: too few, the whole call; too many, from the first extra argument
  to the last one.
- Severity: error.
- Message: `Pass the required argument $name.` (several parameters:
  `Pass the required arguments $a, $b.`), or `Remove the extra argument:
  strlen() accepts at most 1.` (`arguments` for several).

## Fix

None: the missing value or the intended call cannot be guessed.

## Options

None.

## PHP versions

All versions. Builtin callees are checked from PHP 8.0, against the
signature they declare at the configured version.

## Examples

```php
<?php
function greet($name, $greeting) {}
<error descr="Pass the required argument $greeting.">greet('Ada')</error>;
strlen('a', <error descr="Remove the extra argument: strlen() accepts at most 1.">'b'</error>);
```

An override that accepts fewer arguments makes the call valid; extra
arguments to the project's own functions are legal:

```php
<?php
class Renderer { public function render($template, $data) {} }
class Plain extends Renderer { public function render($template, $data = []) {} }
function show(Renderer $r) { $r->render('home'); }
function sum() { return array_sum(func_get_args()); }
sum(1, 2, 3);
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies.
These conditions are independently specified and examples are original.
