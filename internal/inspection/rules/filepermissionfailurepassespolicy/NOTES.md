# FilePermissionFailurePassesPolicy implementation notes

## Budgets and limits

Joined producer retention: when several assignments reach a use, ordinary
producer sets keep at most 64 expressions. Past that, finite resolved native
contracts keep one real representative per false/null guard state;
argument-sensitive contracts keep each producer individually. Origin storage
counts against the scope's retained-state budget; when that budget runs out the
scope becomes unknown instead of inventing a producer or a guard.
