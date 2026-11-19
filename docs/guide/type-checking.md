# Type checking

custos can check types in the same run as its other rules. Type checking is
off by default. Turn it on with `typeCheck` in `custos.json`:

```json
{
  "typeCheck": "documented",
  "rules": {
    "PossiblyNullArgument": { "enabled": false }
  }
}
```

## Levels

| Level | What it reports | Rules |
|---|---|---|
| `off` | Nothing beyond the rules you enable yourself. This is the default. | |
| `proven` | Mistakes that native declarations and the code alone prove. | `InvalidArgumentType`, `InvalidReturnType`, `InvalidPropertyAssignment`, `ArgumentCountMismatch`, `UndefinedFunction`, `UndefinedClass`, `UndefinedMethod`, `UndefinedProperty`, `ImpossibleTypeComparison`, `MemberAccessOnNonObject`, `ForeachOverNonIterable`, `InvalidOperandType` |
| `documented` | Everything in `proven`, plus values that break a documented contract or may be null. | `DocumentedArgumentType`, `DocumentedReturnType`, `DocumentedPropertyAssignment`, `DocumentedMemberAccessOnNonObject`, `DocumentedForeachOverNonIterable`, `DocumentedOperandType`, `PossiblyNullAccess`, `PossiblyNullArgument`, `RedundantTypeCheck` |

A rule's own `enabled` setting always wins over the level, and `--rule` still
selects exactly the rules it names. A finding is only reported when the types
settle it: unknown types, `mixed` and unresolved classes stay silent.

## Declaration formats

custos reads these declarations, whichever tool they were written for.

### PHPDoc, PHPStan and Psalm tags

| Purpose | Tags |
|---|---|
| Parameters, returns, variables | `@param`, `@return`, `@var`, each also with the `@phpstan-` and `@psalm-` prefix |
| Output parameters | `@param-out`, `@phpstan-param-out`, `@psalm-param-out` |
| Templates | `@template`, `@template-covariant`, `@template-contravariant`, each also with the `@phpstan-` and `@psalm-` prefix |
| Inheritance | `@extends`, `@implements`, `@use`, and their `@template-`, `@phpstan-` and `@psalm-` forms |
| Type aliases | `@phpstan-type`, `@psalm-type`, `@phpstan-import-type`, `@psalm-import-type` |
| Assertions | `@phpstan-assert`, `@psalm-assert`, and their `-if-true` and `-if-false` forms |
| Magic members | `@method`, `@property`, `@property-read`, `@property-write`, each also with the `@phpstan-` and `@psalm-` prefix |
| Mixins | `@mixin`, `@phpstan-mixin`, `@psalm-mixin`: members of the named class, or of the class a template stands for, are typed as members of this one |
| Receiver after a call | `@phpstan-self-out`, `@phpstan-this-out`, `@psalm-self-out`, `@psalm-this-out` |

When several tags describe the same thing, `@phpstan-` wins over `@psalm-`,
which wins over the plain tag.

### Type syntax

| Kind | Examples |
|---|---|
| Unions, intersections, nullable | `int\|string`, `A&B`, `?int` |
| Literals | `'asc'`, `42`, `0x1F`, `1_000` |
| Refined scalars | `positive-int`, `int<0, 10>`, `non-empty-string`, `numeric-string`, `int-mask<1, 2, 4>`, `int-mask-of<Foo::FLAG_*>` |
| Arrays | `list<int>`, `non-empty-array<string, int>`, `int[]` |
| Array shapes | `array{id: int, name?: string}`, `array{int, string}`, `array{id: int, ...}` |
| Object shapes | `object{id: int, name?: string}` |
| Callables | `callable(int, string=): bool`, `Closure(int &$x, string ...$rest): void` |
| Generics | `Collection<int, User>`, `class-string<User>`, `enum-string<Suit>` |
| Key and value types | `key-of<Foo::MAP>`, `value-of<Foo::MAP>`, `Foo::BAR_*` |
| Conditional returns | `($x is int ? A : B)` |

### PhpStorm attributes

| Attribute | Effect |
|---|---|
| `#[ArrayShape]`, `#[ObjectShape]` | Describe the array or object member of the type. Their keys count as optional and the shape as open, because PhpStorm's syntax cannot say otherwise. |
| `#[ExpectedValues]` | Limits an int or string parameter or property to the listed literals and class constants, the flags it lists, or the constants of the class named by `valuesFromClass` or `flagsFromClass`. |
| `#[NoReturn]` | Without arguments, the function never returns. With arguments it names only some calls, and is ignored. |
| `#[LanguageLevelTypeAware]` | Gives a parameter or return type per PHP version. |

A plain `@param array` or `@return array` is refined by a shape attribute. A
tag that already has a shape keeps priority over the attribute.

## What is compared

| Comparison | Reported when |
|---|---|
| Generic arguments | An argument has nothing in common with the expected one, following the template's declared variance, including through `@extends` and `@implements`. Arguments count when a declaration states them: an unmodified parameter, a property of `$this`, or a function or method whose documented return mentions no template. |
| Array shapes | A required key is missing, a required key's value has another type, or a sealed shape does not list a required key. |
| Array keys | A non-empty array of int keys is passed where only string keys are documented. String keys never decide, since PHP stores numeric string keys as ints. A key the code proved present, such as with `isset($map[0])`, overrides a key type it contradicts. |
| Object shapes | A property required on both sides has another type. |
| Callables | The return type or a parameter type has nothing in common, or the callable requires more arguments than it receives. |

## Limits

- `#[ExpectedValues]` counts a constant only when it is declared with a
  literal value, and ignores `static::` constants, which a subclass may
  override.
- A method reached through `@mixin` that returns `static` or `$this` has an
  unknown type: what it returns depends on how the class forwards the call.
- `#[Pure]`, `#[Immutable]` and `@throws` do not affect any check.
- A closure is compared by its written signature: what its body was inferred
  to return is not checked against a callable type.
- Extra arguments to project functions are reported only with the
  `strictUserArgs` option of `ArgumentCountMismatch`.
