# UndefinedMethod implementation notes

## Shared check

The rule shares its lookup with UndefinedProperty
(`semanticquery/undefined_members.go`); only the member test (a method,
`__call`, `__callStatic` or a trait alias) is its own:

1. `MemberReceiverClasses` takes the receiver's classes from
   `infer.Env.ProvenTypeOf` (null removed; every atom must be a class;
   `$this` in closures and arrow functions is unknown there), and
   `StaticMemberClasses` those of a static call from `StaticCallClass`,
   skipping `self`, `static` and `parent` in traits and closures;
2. `memberKnown` requires `HierarchyResolved`, `AncestorsComplete`, and a
   single, unconditional declaration (`index.Class.Conditional`) of every
   ancestor at the target version;
3. `MissingMember` walks `index.Ancestors` of each class, then of every
   class `index.Extending` lists unless the lookup is exact (a class name,
   `self`, `parent`). The rule reads the members maps directly rather than
   `FindMethod`, so a member declared for another PHP version, a `@method`
   tag, a trait alias or a trait conflict all count as declared;
4. `index.Class.Mixin` and the `dynamicBuiltins` list stop the search
   as if the member were found;
5. `MemberGuarded` looks for `method_exists`, `is_callable` and
   `property_exists` with the name as a literal in enclosing conditions,
   then asks `NativeEarlierCallProves` for earlier checks.

The index flags `Mixin` and `Conditional` were added for these rules; the
stubs are decoded without them (no stub class carries `@mixin`, and stub
classes declared twice are already counted by `ClassDecls`).

## Subclasses and implementations

Without step 3's walk through `Extending`, laravel gave 100 findings and
composer 8, all on interface or base-class receivers whose implementation
declares the member: the code narrows the type through PHPDoc (`@var`,
`@param` of a closure), through an `instanceof` on `$this` the inference
does not apply, through `class_uses_recursive()`, or a builtin returns a
subclass of its declared type (`PharData` entries are `PharFileInfo`).
The calls sampled all work, so the rule stays silent whenever any known
subtype declares the member.

## Corpus triage

The corpora (guzzle, composer, laravel) give no finding. Composer's ninth
finding before the triage, on `RecursiveIteratorIterator`, led to
`dynamicBuiltins` listing it.
