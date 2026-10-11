<?php
// Guards that do not prove the value is an array at the read.
function reassignedInBranch() { $e = DateTime::getLastErrors(); if ($e) { $e = DateTime::getLastErrors(); return <warning descr="Handle the false result before reading parsing diagnostics.">$e['warning_count']</warning>; } return 0; }
function reassignedAfterExit() { $e = DateTime::getLastErrors(); if (!$e) { return 0; } $e = DateTime::getLastErrors(); return <warning descr="Handle the false result before reading parsing diagnostics.">$e['warning_count']</warning>; }
function reassignedAfterAssign() { if ($e = DateTime::getLastErrors()) { $e = DateTime::getLastErrors(); return <warning descr="Handle the false result before reading parsing diagnostics.">$e['warning_count']</warning>; } return 0; }
function wrongBranch() { $e = DateTime::getLastErrors(); if ($e) { echo 1; } else { return <warning descr="Handle the false result before reading parsing diagnostics.">$e['warning_count']</warning>; } }
function wrongOperator() { $e = DateTime::getLastErrors(); return $e || <warning descr="Handle the false result before reading parsing diagnostics.">$e['warning_count']</warning>; }
function wrongOperand() { $e = DateTime::getLastErrors(); return <warning descr="Handle the false result before reading parsing diagnostics.">$e['warning_count']</warning> && $e; }
function wrongArm() { $e = DateTime::getLastErrors(); return $e ? 0 : <warning descr="Handle the false result before reading parsing diagnostics.">$e['warning_count']</warning>; }
function ternaryCondition() { $e = DateTime::getLastErrors(); return <warning descr="Handle the false result before reading parsing diagnostics.">$e['warning_count']</warning> ? 1 : 0; }
function weakAnd($x) { $e = DateTime::getLastErrors(); return ($e && $x) ? 1 : <warning descr="Handle the false result before reading parsing diagnostics.">$e['warning_count']</warning>; }
function weakOr($x) { $e = DateTime::getLastErrors(); return ($e || $x) ? <warning descr="Handle the false result before reading parsing diagnostics.">$e['warning_count']</warning> : 0; }
function fallsThrough() { $e = DateTime::getLastErrors(); if (!$e) { echo 1; } return <warning descr="Handle the false result before reading parsing diagnostics.">$e['warning_count']</warning>; }
function exitWithElse() { $e = DateTime::getLastErrors(); if (!$e) { return 0; } else { echo 1; } echo 2; return <warning descr="Handle the false result before reading parsing diagnostics.">$e['warning_count']</warning>; }
function issetWhole() { $e = DateTime::getLastErrors(); if (isset($e)) { return <warning descr="Handle the false result before reading parsing diagnostics.">$e['warning_count']</warning>; } return 0; }
function issetNegated() { $e = DateTime::getLastErrors(); if (!isset($e['errors'])) { return <warning descr="Handle the false result before reading parsing diagnostics.">$e['warning_count']</warning>; } return 0; }
function emptyTrue() { $e = DateTime::getLastErrors(); if (empty($e)) { return <warning descr="Handle the false result before reading parsing diagnostics.">$e['warning_count']</warning>; } return 0; }
function otherArray($x) { $e = DateTime::getLastErrors(); if (is_array($x)) { return <warning descr="Handle the false result before reading parsing diagnostics.">$e['warning_count']</warning>; } return 0; }
function notTrue() { $e = DateTime::getLastErrors(); if ($e === true) { return 0; } return <warning descr="Handle the false result before reading parsing diagnostics.">$e['warning_count']</warning>; }
function ordered() { $e = DateTime::getLastErrors(); return $e < 3 && <warning descr="Handle the false result before reading parsing diagnostics.">$e['warning_count']</warning>; }
function negatedOther() { $e = DateTime::getLastErrors(); return -$e && <warning descr="Handle the false result before reading parsing diagnostics.">$e['warning_count']</warning>; }
function elseifCondition($x) { $e = DateTime::getLastErrors(); if ($x) { return 1; } elseif (<warning descr="Handle the false result before reading parsing diagnostics.">$e['warning_count']</warning>) { return 2; } elseif ($e === false) { return 3; } return 0; }
function elseifWrong($x) { $e = DateTime::getLastErrors(); if ($x) { return 1; } elseif (!$e) { return <warning descr="Handle the false result before reading parsing diagnostics.">$e['warning_count']</warning>; } return 0; }
function ifCondition() { $e = DateTime::getLastErrors(); if (<warning descr="Handle the false result before reading parsing diagnostics.">$e['warning_count']</warning>) { return 1; } return 0; }
function indexRead() { $e = DateTime::getLastErrors(); return $e[<warning descr="Handle the false result before reading parsing diagnostics.">$e['warning_count']</warning>] ?? 0; }
function fallback() { $e = DateTime::getLastErrors(); return 0 ?? <warning descr="Handle the false result before reading parsing diagnostics.">$e['warning_count']</warning>; }
function direct() { return <warning descr="Handle the false result before reading parsing diagnostics.">DateTime::getLastErrors()['warning_count']</warning>; }
function unbraced() { $e = DateTime::getLastErrors(); if (rand()) return <warning descr="Handle the false result before reading parsing diagnostics.">$e['warning_count']</warning>; return 0; }
