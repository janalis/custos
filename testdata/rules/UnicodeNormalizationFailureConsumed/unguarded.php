<?php
function truthy(string $i): int { $s = normalizer_normalize($i); if (!$s) { return 0; } return <warning descr="Check normalization before consuming text.">strlen($s)</warning>; }
function loose(string $i): int { $s = normalizer_normalize($i); if ($s == false) { return 0; } return <warning descr="Check normalization before consuming text.">strlen($s)</warning>; }
function truthyTernary(string $i): int { $s = normalizer_normalize($i); return $s ? <warning descr="Check normalization before consuming text.">strlen($s)</warning> : 0; }
function shortTernary(string $i): int { $s = normalizer_normalize($i); return $i ?: <warning descr="Check normalization before consuming text.">strlen($s)</warning>; }
function asserted(string $i): int { $s = normalizer_normalize($i); assert($s !== false); return <warning descr="Check normalization before consuming text.">strlen($s)</warning>; }
function wrongBranch(string $i): int { $s = normalizer_normalize($i); return $s !== false ? 0 : <warning descr="Check normalization before consuming text.">strlen($s)</warning>; }
function wrongSide(string $i): bool { $s = normalizer_normalize($i); return <warning descr="Check normalization before consuming text.">strlen($s)</warning> > 3 && $s !== false; }
function wrongOperator(string $i): bool { $s = normalizer_normalize($i); return $s !== false || <warning descr="Check normalization before consuming text.">strlen($s)</warning> > 3; }
function otherOperator(string $i): int { $s = normalizer_normalize($i); return ($s !== false) + <warning descr="Check normalization before consuming text.">strlen($s)</warning>; }
function otherVariable(string $i, $o): int { $s = normalizer_normalize($i); if ($o === false) { return 0; } return <warning descr="Check normalization before consuming text.">strlen($s)</warning>; }
function otherTest(string $i, $o): int { $s = normalizer_normalize($i); if (!is_string($o) || !is_string() || !is_int($s) || $s === null || $s > false) { return 0; } return <warning descr="Check normalization before consuming text.">strlen($s)</warning>; }
function rebound(string $i): int { $s = normalizer_normalize($i); if ($s === false) { return 0; } $s = normalizer_normalize($i . 'a'); return <warning descr="Check normalization before consuming text.">strlen($s)</warning>; }
function earlyCheck(string $i): int { $s = ''; if (($s = $i) === false) { return 0; } $s = normalizer_normalize($i); return <warning descr="Check normalization before consuming text.">strlen($s)</warning>; }
function fallsThrough(string $i): int { $s = normalizer_normalize($i); if ($s === false) { $i = ''; } return <warning descr="Check normalization before consuming text.">strlen($s)</warning>; }
function unrelatedExit(string $i): int { $s = normalizer_normalize($i); $i === '' or exit; $i; return <warning descr="Check normalization before consuming text.">strlen($s)</warning>; }
function inCondition(string $i): int { $s = normalizer_normalize($i); if (<warning descr="Check normalization before consuming text.">strlen($s)</warning> > 3) { return 1; } elseif (<warning descr="Check normalization before consuming text.">strlen($s)</warning> > 2) { return 2; } return 0; }
function elseIfBody(string $i): int { $s = normalizer_normalize($i); if ($i === '') { return 0; } elseif ($i === 'a') { return <warning descr="Check normalization before consuming text.">strlen($s)</warning>; } elseif ($s === false) { return 1; } return 2; }
function matchOther(string $i): int { $s = normalizer_normalize($i); return match ($i) { '' => 0, default => <warning descr="Check normalization before consuming text.">strlen($s)</warning> }; }
function matchFalse(string $i): int { $s = normalizer_normalize($i); return match (false) { $s === false => 0, default => <warning descr="Check normalization before consuming text.">strlen($s)</warning> }; }
function matchFirst(string $i): int { $s = normalizer_normalize($i); return match (true) { $i === '', $s !== false => <warning descr="Check normalization before consuming text.">strlen($s)</warning>, default => 0 }; }
function matchLater(string $i): int { $s = normalizer_normalize($i); return match (true) { $i === '' => <warning descr="Check normalization before consuming text.">strlen($s)</warning>, $s === false => 0, default => 1 }; }
function matchValue(string $i): int { $s = normalizer_normalize($i); return match ($s) { '' => 0, true => 1, default => <warning descr="Check normalization before consuming text.">strlen($s)</warning> }; }
function direct(string $i): int { return normalizer_normalize($i) !== false ? <warning descr="Check normalization before consuming text.">strlen(normalizer_normalize($i))</warning> : 0; }
function thirdBranch(string $i): int { $s = normalizer_normalize($i); if ($i === '') { return 0; } elseif ($i === 'a') { return 1; } elseif ($i === 'b') { return <warning descr="Check normalization before consuming text.">strlen($s)</warning>; } return 2; }
