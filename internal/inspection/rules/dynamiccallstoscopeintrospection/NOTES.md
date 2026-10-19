# DynamicCallsToScopeIntrospection implementation notes

## Known gaps

- Value discovery (`semanticquery.LocalValuesReaching`) takes only `b` from a
  short ternary `a ?: b`. Taking the union of `a` and `b`, as the shared
  `DiscoverValues` does, would find `'extract' ?: $other`.
