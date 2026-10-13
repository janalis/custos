<?php
class Target {}

// Binding null to a static closure succeeds, so a nullable object proves nothing.
function nullableObject(?Target $o) { $c = static fn() => 1; $b = $c->bindTo($o); return $b(); }

function truthy(Target $o) { $c = static fn() => 1; $b = $c->bindTo($o); if ($b) { return $b(); } return 0; }
function instance(Target $o) { $c = static fn() => 1; $b = $c->bindTo($o); if ($b instanceof Closure) { return $b(); } return 0; }
function ternary(Target $o) { $c = static fn() => 1; $b = $c->bindTo($o); return $b ? $b() : 0; }
function ternaryElse(Target $o) { $c = static fn() => 1; $b = $c->bindTo($o); return $b === null ? 0 : $b(); }
function shortCircuit(Target $o) { $c = static fn() => 1; $b = $c->bindTo($o); return $b !== null && $b(); }
function shortCircuitOr(Target $o) { $c = static fn() => 1; $b = $c->bindTo($o); return !$b or $b(); }
function isNull(Target $o) { $c = static fn() => 1; $b = $c->bindTo($o); if (is_null($b)) { return 0; } return $b(); }
function callable_(Target $o) { $c = static fn() => 1; $b = $c->bindTo($o); return is_callable($b) ? $b() : 0; }
function issetGuard(Target $o) { $c = static fn() => 1; $b = $c->bindTo($o); return isset($b) ? $b() : 0; }
function emptyGuard(Target $o) { $c = static fn() => 1; $b = $c->bindTo($o); if (empty($b)) { return 0; } return $b(); }
function conjunction(Target $o, $flag) { $c = static fn() => 1; $b = $c->bindTo($o); if ($flag && null != $b) { return $b(); } return 0; }
function disjunction(Target $o, $flag) { $c = static fn() => 1; $b = $c->bindTo($o); if ($flag || $b == null) { return 0; } return $b(); }
function earlierGuard(Target $o) { $c = static fn() => 1; $b = $c->bindTo($o); if (!$b) { throw new Exception(); } echo 'bound'; return $b(); }
function elseBranch(Target $o) { $c = static fn() => 1; $b = $c->bindTo($o); if ($b === null) { return 0; } else { return $b(); } }
function elseIfBranch(Target $o, $flag) { $c = static fn() => 1; $b = $c->bindTo($o); if ($flag) { return 0; } elseif ($b) { return $b(); } return 0; }
function afterFailedBranch(Target $o, $flag) { $c = static fn() => 1; $b = $c->bindTo($o); if (!$b) { return 0; } elseif ($flag) { return $b(); } elseif ($b()) { return 1; } return 0; }
function assigned(Target $o) { $c = static fn() => 1; if ($b = $c->bindTo($o)) { return $b(); } return 0; }
