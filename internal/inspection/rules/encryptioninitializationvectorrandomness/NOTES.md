# EncryptionInitializationVectorRandomness implementation notes

## Shared helpers

- Reaching values of a local IV argument come from
  `semanticquery.LocalValuesReaching` (plain `=` assignments that reach the
  call, plus the parameter default when the entry value reaches it).
- `ivReplacedValues` then drops placeholders that `flowquery.ReachingAssignmentsIn`
  keeps although every path replaces them before the call (a `try` whose
  catches all assign or leave, a `do … while` body, an `if` testing the
  variable for the placeholder): see E6 in the spec.
