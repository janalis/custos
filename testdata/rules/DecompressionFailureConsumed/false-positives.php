<?php
function ternaryFailure($b) { $s = gzdecode($b); return $s === false ? 0 : strlen($s); }
function ternarySuccess($b) { $s = gzdecode($b); return $s !== false ? strlen($s) : 0; }
function shortTernary($b) { $s = gzinflate($b); return $s === false ?: strlen($s); }
function conjunction($b) { $s = gzdecode($b); return $s !== false && strlen($s) > 3; }
function disjunction($b) { $s = gzdecode($b); return $s === false || strlen($s) === 0; }
function wordOperators($b) { $s = gzuncompress($b); return (false !== $s and strlen($s) > 3) or (false === $s or strlen($s) > 9); }
function assignedFailure($b) { if (($s = gzdecode($b)) === false) { return 0; } return strlen($s); }
function assignedSuccess($b) { if (($s = gzdecode($b)) !== false) { return strlen($s); } return 0; }
function silencedAssignment($b) { if (($s = @gzdecode($b)) === false) { return 0; } return strtoupper($s); }
function compoundFailure($b) { $s = gzdecode($b); if ($s === false || strlen($s) < 4) { return 0; } return strlen($s); }
function compoundSuccess($b, $ok) { $s = gzdecode($b); if ($ok && $s !== false) { return strlen($s); } return 0; }
function negated($b) { $s = gzdecode($b); if (!($s === false)) { return strlen($s); } return 0; }
function laterArms($b, $x) { $s = gzdecode($b); if ($s === false) { return 0; } elseif ($x) { return strlen($s); } else { return strtoupper($s); } }
function ownArm($b, $x) { $s = gzdecode($b); if ($x) { return 0; } elseif ($s !== false) { return strlen($s); } return 1; }
function whileLoop($c, $b) { while (($s = inflate_add($c, $b)) !== false) { return strlen($s); } return -1; }
function forLoop($c, $b) { for ($i = 0; $i < 3, ($s = inflate_add($c, $b)) !== false; $i++) { return strlen($s); } return -1; }
function silencedGuarded($b) { $s = @gzdecode($b); if ($s === false) { return 0; } return strlen($s); }
function assignedArm($b, $x) { if ($x) { return 0; } elseif (($s = gzdecode($b)) !== false) { return strlen($s); } return 1; }
function assignedArmTest($b, $x) { if ($x) { return 0; } elseif (($s = gzdecode($b)) !== false && strlen($s) > 3) { return 2; } return 1; }
function afterFailedArm($b, $x) { if ($x) { return 0; } elseif (($s = (gzdecode($b))) === false) { return 2; } elseif ($b) { return strtoupper($s); } else { return strlen($s); } }
function exitThenWork($b) { if (($s = gzdecode($b)) === false) { return 0; } $n = 1; return strlen($s) + $n; }
function exitThenNested($b, $x) { $s = gzdecode($b); if (false === $s) { throw new RuntimeException('bad'); } $n = 1; if ($x) return strlen($s) + $n; return 0; }
