<?php
// Zero is told apart before the comparison, in its condition, or in the
// statement the comparison governs.
function earlyExit($pid) { $r = pcntl_waitpid($pid, $s, WNOHANG); if ($r === 0) { return 'running'; } if ($r !== -1) { return 'exited'; } return 'error'; }
function sibling($pid) { $r = pcntl_waitpid($pid, $s, WNOHANG); if ($r !== -1 && $r !== 0) { return 'exited'; } return $r === 0 ? 'running' : 'error'; }
function inside($pid) { $r = pcntl_waitpid($pid, $s, WNOHANG); if ($r != -1) { if ($r == 0) { return 'running'; } return 'exited'; } return 'error'; }
function negated($pid) { $r = pcntl_waitpid($pid, $s, WNOHANG); if (!$r) { return; } if ($r >= 0) { echo $s; } }
function enclosing($pid) { $r = pcntl_waitpid($pid, $s, WNOHANG); if (($r)) { if ($r != -1) { echo $s; } } }
function chained($pid, $x) { $r = pcntl_waitpid($pid, $s, WNOHANG); if ($x) { return; } elseif ($r) { echo $r !== -1; } }
function ternary($pid) { $r = pcntl_waitpid($pid, $s, WNOHANG); return $r ? $r !== -1 : 0; }
function switched($pid) { $r = pcntl_waitpid($pid, $s, WNOHANG); switch ($r) { case 0: return 0; default: return $r !== -1; } }
function yodaPositive($pid) { $r = pcntl_waitpid($pid, $s, WNOHANG); return 0 < $r && $r !== -1; }
function yodaOne($pid) { $r = pcntl_waitpid($pid, $s, WNOHANG); return 1 <= $r && $r !== -1; }
function belowOne($pid) { $r = pcntl_waitpid($pid, $s, WNOHANG); if ($r < 1) { return; } return $r !== -1; }
function truthy($pid) { $r = pcntl_waitpid($pid, $s, WNOHANG); return $r && $r !== -1; }
function either($pid) { $r = pcntl_waitpid($pid, $s, WNOHANG); if ($r <= 0 || $r > 100000) { return; } return $r !== -1; }
