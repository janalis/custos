# CallableMethodValidity implementation notes

## Value discovery

The spec's value discovery (D3) is `semanticquery.DiscoverValues`, the
index-aware variant of `flowquery.PossibleValues`. Several other rules reuse
it, so changes there are shared. It collects the candidate value expressions
of an expression, visiting each once with parentheses stripped:

- ternary: both branches (`a ?: b`: `a` and `b`); `a ?? b`: both operands;
- variable: the parameter default and every plain `=` assignment (innermost
  value of a chain) in the enclosing function-like scope, nested closures
  included; nothing in top-level code;
- property fetch (`$o->p`, `X::$p`): resolved through the receiver type and
  the index; the declared default (unless its text ends with the property
  name) plus assignments to an equivalent fetch in the current scope and in
  the declaring class's constructor (declaration in this file only);
- class constant: resolved through the index (inheritance aware); its value
  expression is discovered (declaration in this file only);
- global constant other than `true`/`false`/`null`: resolved through the
  index. A `define()`/`const` declared in this file contributes its value
  expression as is (not discovered further); one declared elsewhere (project,
  stubs) contributes the constant reference itself; an unresolved one
  contributes nothing;
- anything else is a value itself.

A variable expanded at any depth that is also incremented, decremented or
compound-assigned in its scope (`UnstableVariableIn`) makes the whole result
unknown: `DiscoverValues` returns nil and `DiscoverValuesKnown` reports
`known == false`. Consumers for which an empty result is not already "no
report" must use `DiscoverValuesKnown`.
