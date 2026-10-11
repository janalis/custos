<?php
// Tests that do not reject a failed lookup of the value being masked.
function otherValue($f) { $p = fileperms($f); if (!$p) { return false; } $p = fileperms($f . 'x'); if (<warning descr="Reject permission lookup failure before checking bits.">($p & 0222) === 0</warning>) { return true; } return false; }
function otherVariable($f, $q) { $p = fileperms($f); if (!$q) { return false; } if (<warning descr="Reject permission lookup failure before checking bits.">($p & 0222) === 0</warning>) { return true; } return false; }
function continues($f) { $p = fileperms($f); if (!$p) { error_log($f); } if (<warning descr="Reject permission lookup failure before checking bits.">($p & 0222) === 0</warning>) { return true; } return false; }
function guardWithElse($f) { $p = fileperms($f); if (!$p) { return false; } else { echo $f; } if (<warning descr="Reject permission lookup failure before checking bits.">($p & 0222) === 0</warning>) { return true; } return false; }
function falsyElse($f) { $p = fileperms($f); if ($p) { return 1; } else { if (<warning descr="Reject permission lookup failure before checking bits.">($p & 0222) === 0</warning>) { return true; } } return false; }
function unrelatedBranch($f) { $p = fileperms($f); if (is_int($f)) { if (<warning descr="Reject permission lookup failure before checking bits.">($p & 0222) === 0</warning>) { return true; } } return false; }
function partialGuard($f) { $p = fileperms($f); if (!$p && $f) { return false; } if (<warning descr="Reject permission lookup failure before checking bits.">($p & 0222) === 0</warning>) { return true; } return false; }
function weakTests($f) { $p = fileperms($f); if (-$p || is_string($p) || is_int($p, 1) || is_int(...) || is_int(...[$p]) || $p === null || $p < 1 || is_int($p) || !is_bool($p) || !empty($p) || $p !== false) { return false; } if (<warning descr="Reject permission lookup failure before checking bits.">($p & 0222) === 0</warning>) { return true; } return false; }
function invertedTests($f) { $p = fileperms($f); if (empty($p) && !is_int($p) && $p === false && $p == false && is_bool($p) && !$p) { if (<warning descr="Reject permission lookup failure before checking bits.">($p & 0222) === 0</warning>) { return true; } } return false; }
function existenceOnly($f) { if (!file_exists($f)) { return false; } $p = fileperms($f); if (<warning descr="Reject permission lookup failure before checking bits.">($p & 0222) === 0</warning>) { return true; } return false; }
// Not examined: D1 covers only a comparison that is the whole condition of an if.
function falsyElseIf($f) { $p = fileperms($f); if ($p) { return 1; } elseif ($f) { return 2; } elseif (($p & 0222) === 0) { return true; } return false; }
function shortCircuitOther($f) { $p = fileperms($f); if ($f && ($p & 0222) === 0) { return true; } return false; }
