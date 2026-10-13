<?php
// Probes that PHP evaluates silently on false.
function coalesce() { $e = DateTime::getLastErrors(); return $e['warning_count'] ?? 0; }
function nestedCoalesce() { $e = DateTime::getLastErrors(); return ($e['warnings'])[3] ?? ''; }
function viaIsset() { $e = DateTime::getLastErrors(); return isset($e['warnings']) ? 1 : 0; }
function viaEmpty() { $e = DateTime::getLastErrors(); return empty($e['errors']); }

// Short-circuit and ternary guards.
function andGuard() { $e = DateTime::getLastErrors(); return $e && $e['warning_count']; }
function wordAnd() { $e = DateTime::getLastErrors(); return ($e and $e['warning_count']); }
function chainedAnd() { $e = DateTime::getLastErrors(); return $e && $e['warning_count'] && $e['error_count']; }
function orGuard() { $e = DateTime::getLastErrors(); return !$e || $e['warning_count']; }
function wordOr() { $e = DateTime::getLastErrors(); return ($e === false or $e['warning_count'] === 0); }
function strictAnd() { $e = DateTime::getLastErrors(); return $e !== false && $e['warning_count'] > 0; }
function looseAnd() { $e = DateTime::getLastErrors(); return false != $e && $e['warning_count'] > 0; }
function strictOr() { $e = DateTime::getLastErrors(); return $e === false || $e['warning_count'] === 0; }
function ternary() { $e = DateTime::getLastErrors(); return $e ? $e['warning_count'] : 0; }
function negatedTernary() { $e = DateTime::getLastErrors(); return !$e ? 0 : $e['warning_count']; }
function compoundTernary() { $e = DateTime::getLastErrors(); return ($e && $e['error_count'] === 0) ? $e['warning_count'] : 0; }
function issetAnd() { $e = DateTime::getLastErrors(); return isset($e['warnings']) && $e['warnings'][0]; }

// Branch guards.
function compoundIf() { $e = DateTime::getLastErrors(); if ($e !== false && $e['warning_count'] > 0) { return 1; } return 0; }
function isArray() { $e = DateTime::getLastErrors(); if (is_array($e)) { return $e['warning_count']; } return 0; }
function notEmpty() { $e = DateTime::getLastErrors(); if (!empty($e)) { return $e['warning_count']; } return 0; }
function assigned() { if ($e = DateTime::getLastErrors()) { return $e['warning_count']; } return 0; }
function assignedStrict() { if (($e = DateTime::getLastErrors()) !== false) { return $e['warning_count']; } return 0; }
function elseBranch() { $e = DateTime::getLastErrors(); if (!$e) { return 0; } else { return $e['warning_count']; } }
function elseifBody($x) { $e = DateTime::getLastErrors(); if ($x) { return 1; } elseif ($e) { return $e['warning_count']; } return 0; }
function elseifCond() { $e = DateTime::getLastErrors(); if (!$e) { return 1; } elseif ($e['warning_count']) { return 2; } return 0; }
function afterElseif($x) { $e = DateTime::getLastErrors(); if ($x) { return 1; } elseif ($e === false) { return 2; } else { return $e['error_count']; } }
function afterNegatedElseif($x) { $e = DateTime::getLastErrors(); if ($x) { return 1; } elseif (!$e) { return 2; } else { return $e['error_count']; } }

// Early exits.
function earlyReturn() { $e = DateTime::getLastErrors(); if (!$e) { return 0; } return $e['warning_count']; }
function earlyThrow() { $e = DateTime::getLastErrors(); if (empty($e)) { throw new RuntimeException('none'); } echo 1; return $e['warning_count']; }
function earlyLoose($x) { $e = DateTime::getLastErrors(); if ($e == false || $x) { return 0; } return $e['warning_count']; }
function earlyContinue(array $rows) { $e = DateTime::getLastErrors(); foreach ($rows as $row) { if (!$e) { continue; } echo $e['warning_count'], $row; } }
