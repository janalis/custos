<?php
// Each function tests the lookup result before reading its bits.
function falsyGuard($f) { $p = fileperms($f); if (!$p) { return false; } if (($p & 0222) === 0) { return true; } return false; }
function emptyGuard($f) { $p = fileperms($f); if (empty($p)) { return false; } if (($p & 0222) === 0) { return true; } return false; }
function intGuard($f) { $p = fileperms($f); if (!is_int($p)) { return false; } if (($p & 0222) === 0) { return true; } return false; }
function boolGuard($f) { $p = fileperms($f); if (is_bool($p) || $f === '') { return false; } echo $f; if (($p & 0222) === 0) { return true; } return false; }
function looseGuard($f) { $p = fileperms($f); if ($p == FALSE) { return false; } if (($p & 0222) === 0) { return true; } return false; }
function yodaGuard($f) { $p = fileperms($f); if (false === $p || $f === '') { throw new RuntimeException($f); } if (($p & 0222) === 0) { return true; } return false; }
function truthyBranch($f) { $p = fileperms($f); if ($p) { if (($p & 0222) === 0) { return true; } } return false; }
function intBranch($f) { $p = fileperms($f); if ($f !== '' && is_int($p)) { if (($p & 0222) === 0) { return true; } } return false; }
function looseBranch($f) { $p = fileperms($f); if ($p != false) { if (($p & 0222) === 0) { return true; } } return false; }
function shortCircuit($f) { $p = fileperms($f); if ($p && ($p & 0222) === 0) { return true; } return false; }
function alternative($f) { $p = fileperms($f); if ($f === '' || !$p || ($p & 0222) !== 0) { return false; } return true; }
function elseArm($f) { $p = fileperms($f); if (!$p) { return false; } else { if (($p & 0222) === 0) { return true; } } return false; }
function elseIfArm($f) { $p = fileperms($f); if ($f === '') { return 1; } elseif (!is_int($p)) { return false; } elseif (($p & 0222) === 0) { return true; } return false; }
function elseIfBody($f) { $p = fileperms($f); if ($f === '') { return 1; } elseif ($p) { if (($p & 0222) === 0) { return true; } } return false; }
function afterElseIf($f) { $p = fileperms($f); if ($f === '') { return 1; } elseif (!$p) { return false; } else { if (($p & 0222) === 0) { return true; } } return false; }
function orRight($f) { $p = fileperms($f); if (!$p || $f === '') { return false; } if (!$p || ($p & 0222) === 0) { return true; } return false; }
