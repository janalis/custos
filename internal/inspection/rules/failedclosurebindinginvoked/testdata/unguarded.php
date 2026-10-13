<?php
class Target {}

function unrelatedGuard(Target $o, $other) { $c = static fn() => 1; $b = $c->bindTo($o); if ($other) { return <error descr="Check closure binding before invoking it.">$b()</error>; } return 0; }
function wrongBranch(Target $o) { $c = static fn() => 1; $b = $c->bindTo($o); if ($b) { return 0; } return <error descr="Check closure binding before invoking it.">$b()</error>; }
function wrongArm(Target $o) { $c = static fn() => 1; $b = $c->bindTo($o); return $b ? 0 : <error descr="Check closure binding before invoking it.">$b()</error>; }
function wrongOperand(Target $o) { $c = static fn() => 1; $b = $c->bindTo($o); return $b || <error descr="Check closure binding before invoking it.">$b()</error>; }
function nullBranch(Target $o) { $c = static fn() => 1; $b = $c->bindTo($o); if (is_null($b)) { return <error descr="Check closure binding before invoking it.">$b()</error>; } return 0; }
function nonTerminating(Target $o) { $c = static fn() => 1; $b = $c->bindTo($o); if (!$b) { echo 'failed'; } return <error descr="Check closure binding before invoking it.">$b()</error>; }
function otherComparison(Target $o) { $c = static fn() => 1; $b = $c->bindTo($o); if ($b !== false) { return <error descr="Check closure binding before invoking it.">$b()</error>; } return 0; }
function firstCondition(Target $o, $flag) { $c = static fn() => 1; $b = $c->bindTo($o); if (<error descr="Check closure binding before invoking it.">$b()</error>) { return 1; } elseif ($flag) { return 2; } return 0; }
function otherCall(Target $o) { $c = static fn() => 1; $b = $c->bindTo($o); return is_string($b) ? 0 : <error descr="Check closure binding before invoking it.">$b()</error>; }
function concatenated(Target $o) { $c = static fn() => 1; $b = $c->bindTo($o); return 'x' . <error descr="Check closure binding before invoking it.">$b()</error>; }
function nullArm(Target $o) { $c = static fn() => 1; $b = $c->bindTo($o); if ($b === null) { return <error descr="Check closure binding before invoking it.">$b()</error>; } return 0; }
