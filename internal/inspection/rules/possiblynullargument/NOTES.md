# PossiblyNullArgument implementation notes

## Shared check

The rule reports `semanticquery.PossiblyNullArguments`, cached per call.
Arguments are typed first (`possiblyNull`, see the PossiblyNullAccess
notes); only when one may be null is the callee resolved, with the targets
and binding of InvalidArgumentType (`CallTargets`, `bindAll`,
`TrustedBuiltinTypes`). A target counts when its native parameter type
rejects `null` (`types.AllRejected` in the file's acceptance context,
whose `Internal` flag keeps null to a scalar builtin parameter outside
strict mode a Maybe). An argument whose whole proven type is rejected is
left to InvalidArgumentType, so the two never report the same argument.
