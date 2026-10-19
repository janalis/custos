# SlowArrayOperationsInLoop implementation notes

## Changing loop subjects

`forSubjectChange` returns one of three classes for the value measured by a
`for` condition:

- `forSubjectWritten`: the body, the step or a condition visibly changes the
  measured expression, one of its elements, or a part it is reached through
  (root array, object holding a measured property, offset variable). Writes
  are assignments, `list()`/`[]` destructuring, `foreach` key/value targets,
  pushes, increments, `unset`, and by-reference arguments of resolved
  functions. The loop changes its length on purpose: no report.
- `forSubjectMayChange`: the subject is not a variable or property fetch, or
  the loop iterates it by reference, binds a reference to it, passes it to an
  unresolved callee or unpacks it, calls a method on it or on an object it is
  reached through, hands such an object to user code, or runs code that can
  reach a shared subject (a property, or a non-local variable: any call but a
  builtin without callable parameter). A reference to a part anywhere in the
  enclosing scope also lands here. Reported without a fix, since hoisting
  would freeze a changing length.
- otherwise the subject is stable and the fix is offered.

## Fix

When both comparison operands are length calls, only the first (left) one
receives a fix; the second is reported bare so the two edits never conflict.
