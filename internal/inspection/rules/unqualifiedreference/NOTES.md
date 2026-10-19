# UnqualifiedReference implementation notes

## Known gaps

A call inside the global block `namespace { … }` of a multi-namespace file
is reported although it already binds globally. `uqEligible` could skip it
when the governing namespace (`uqGoverningNamespace`) has no name.
