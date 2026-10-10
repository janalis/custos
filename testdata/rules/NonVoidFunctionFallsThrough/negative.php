<?php
function countItems(bool $ready): int { if ($ready) return 7; return 0; }
function literalTruthy(): int { if (1) { return 1; } }
function comparisonTruthy(): int { if (1 === 1) { return 1; } }
function selfComparison($value): int { if ($value === $value) { return 1; } }

function f(): int {1/0;} try {f();} catch (DivisionByZeroError $e) {echo 'expected exception';}
