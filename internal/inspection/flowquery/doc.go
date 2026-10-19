// Package flowquery provides local control-flow and variable-use queries for
// inspections. Queries are syntactic and file-local: they never consult the
// project index, and resolved argument effects are supplied by the caller
// through a mayWrite callback.
//
// The package answers four kinds of question:
//
//   - reachability: [Reachable] decides whether a node can be reached from
//     its scope entry, honoring terminating statements and goto labels;
//   - assignments: [ReachingAssignments] approximates reaching definitions
//     of a local variable, and [UnstableVariableIn], [AssignmentsOfKind] and
//     [CtorAssignsProperty] describe how a variable or property is written;
//   - variable use: [VarAccesses] lists reads and writes of a variable in one
//     scope, treating closure imports and arrow functions as captures;
//   - values: [PossibleValues], [PossibleValuesReaching] and
//     [PossibleValuesComplete] collect the expressions a value may come from.
//
// Results are conservative. An empty value set from [PossibleValues] may mean
// "unknown"; consumers that would report on an empty set use the Known
// variants, which return known == false when a write cannot be followed, and
// stay silent then. [PossibleValuesComplete] is for consumers that need every
// value: when complete is false the candidates must not be relied upon.
// Unsupported constructs, unknown argument bindings and exhausted work bounds
// make a result unknown or incomplete rather than authorize a finding.
package flowquery
