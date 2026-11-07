# UnqualifiedReference implementation notes

## Namespace context

The shared name resolver supplies the effective namespace at a source offset.
An empty namespace is global, whether it comes from a braced global block or a
file without namespace declarations. The governing declaration is still used
for the rule's explicit-import checks.
