<?php
function countItems(bool $ready): int { if ($ready) return 7; return 0; }
function literalTruthy(): int { if (1) { return 1; } }
function comparisonTruthy(): int { if (1 === 1) { return 1; } }
function selfComparison($value): int { if ($value === $value) { return 1; } }

function f(): int {1/0;} try {f();} catch (DivisionByZeroError $e) {echo 'expected exception';}

// Exhaustive sibling branches: the second test is the exact complement of the first.
function signOrZero(int $n): int { if ($n > 0) { return 1; } if ($n <= 0) { return -1; } }
function mirrored($low, $high): int { if ($low < $high) { return 1; } if ($high <= $low) { return 0; } }
function presence($value): int { if ($value === null) { return 0; } echo 'set'; if ($value !== null) { return 1; } }
function boolFlag(bool $flag): int { if ($flag === true) { return 1; } if ($flag === false) { return 0; } }
function nestedComplement($n): int { if ($n == 1) { return 1; } else { if ($n != 1) { return 2; } } }
// Expressions that cannot complete.
function rethrow($e): int { match (true) { default => throw $e }; }
function everyArm($e, $code): int { match ($code) { 1 => throw $e, 2 => throw new LogicException('two') }; }
function assignedThrow($e): int { $unused = throw $e; }
// A call beside other operators may still leave the function.
function callBesideArithmetic(int $x): int { if ($x > 1) { return 1; } exit_now() || $x + 1; }
function evaluated(int $x, string $code): int { if ($x > 1) { return 1; } eval($code); }
