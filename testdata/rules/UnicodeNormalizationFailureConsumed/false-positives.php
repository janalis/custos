<?php
function ternaryElse(string $i): int { $s = normalizer_normalize($i); return $s === false ? 0 : strlen($s); }
function ternaryThen(string $i): int { $s = normalizer_normalize($i); return false !== $s ? strlen($s) : 0; }
function conjunction(string $i): bool { $s = normalizer_normalize($i); return $s !== false && strlen($s) > 3; }
function disjunction(string $i): bool { $s = normalizer_normalize($i); return $s === false || strlen($s) > 3; }
function keywordAnd(string $i): bool { $s = normalizer_normalize($i); return ($s !== false and strlen($s) > 3); }
function assigned(string $i): int { if (($s = normalizer_normalize($i)) === false) { return 0; } return strlen($s); }
function assignedYoda(string $i): int { if (false === ($s = Normalizer::normalize($i))) { return 0; } return strlen($s); }
function typed(string $i): int { $s = normalizer_normalize($i); if (!is_string($s)) { return 0; } return strlen($s); }
function typedBody(string $i): int { $s = normalizer_normalize($i); if (is_string($s)) { return strlen($s); } return 0; }
function typedFalse(string $i): int { $s = normalizer_normalize($i); if (\is_string($s) === false) { return 0; } return strlen($s); }
function typedTrue(string $i): int { $s = normalizer_normalize($i); if (true !== is_string($s)) { throw new RuntimeException('x'); } return strlen($s); }
function matchTrue(string $i): int { $s = normalizer_normalize($i); return match (true) { $s === false => 0, default => strlen($s) }; }
function matchBody(string $i): int { $s = normalizer_normalize($i); return match (true) { $s !== false => strlen($s), default => 0 }; }
function matchSubject(string $i): int { $s = normalizer_normalize($i); return match ($s) { false => 0, default => strlen($s) }; }
function method(string $i): string { $s = Normalizer::normalize($i); return $s === false ? '' : strtoupper($s); }
function unbraced(string $i): int { $s = normalizer_normalize($i); if ($s === false) exit; return strlen($s); }
function orDie(string $i): int { $s = normalizer_normalize($i); $s !== false or die(); return strlen($s); }
function andThrow(string $i): int { $s = normalizer_normalize($i); $s === false && throw new RuntimeException('x'); return strlen($s); }
function later(string $i): int { $s = normalizer_normalize($i); if ($s === false) { return 0; } $n = 1; return strlen($s) + $n; }
function elseBranch(string $i): int { $s = normalizer_normalize($i); if ($s === false) { $n = 0; } else { $n = strlen($s); } return $n; }
function elseIfBranch(string $i): int { $s = normalizer_normalize($i); if ($i === '') { return 0; } elseif ($s !== false) { return strlen($s); } return 1; }
function afterElseIf(string $i): int { $s = normalizer_normalize($i); if ($i === '') { return 0; } elseif ($s === false) { return 1; } elseif ($i === 'a') { return strlen($s); } else { return strlen($s) + 1; } }
function mixedCondition(string $i): int { $s = normalizer_normalize($i); if ($s === false || $i === '') { return 0; } return strlen($s); }
function nestedStatement(string $i): int { $s = normalizer_normalize($i); if ($s === false) { return 0; } if ($i !== '') { return strlen($s); } return 1; }
function withElse(string $i): int { $s = normalizer_normalize($i); if ($s === false) { return 0; } else { $i = ''; } return strlen($s) + strlen($i); }
function typedElse(string $i): int { $s = normalizer_normalize($i); if (!is_string($s)) { $n = 0; } else { $n = strlen($s); } return $n; }
function typedElseIf(string $i): int { $s = normalizer_normalize($i); if ($i === '') { return 0; } elseif (is_string($s)) { return strlen($s); } return 1; }
function typedLastElse(string $i): int { $s = normalizer_normalize($i); if ($i === '') { $n = 0; } elseif (!is_string($s)) { $n = 1; } else { $n = strlen($s); } return $n; }
