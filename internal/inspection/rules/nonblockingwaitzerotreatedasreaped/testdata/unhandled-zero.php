<?php
// Tests that do not separate zero from a reaped PID, concern another wait
// result, or come after the guarded statement leave the finding in place.
function later($pid) { $r = pcntl_waitpid($pid, $s, WNOHANG); if (<warning descr="Require a positive PID from nonblocking wait.">$r !== -1</warning>) { echo $s; } if ($r === 0) { echo 'running'; } }
function unrelated($pid, $x) { $r = pcntl_waitpid($pid, $s, WNOHANG); if ($r < 0) { return; } if ($r === $x) {} if ($x === $r) {} if ($r + 1) {} if (-$r) {} if ($x === 0) {} if (<warning descr="Require a positive PID from nonblocking wait.">$r >= 0</warning>) { echo $s; } }
function otherResult($pid) { $q = pcntl_waitpid($pid, $s, WNOHANG); $r = pcntl_waitpid($pid, $s, WNOHANG); if ($q === 0) { return; } if (<warning descr="Require a positive PID from nonblocking wait.">$r !== -1</warning>) { echo $s; } }
function weak($pid) { $r = pcntl_waitpid($pid, $s, WNOHANG); if ($r <=> 0) {} if (1 === $r) {} if (0 <= $r) {} while ($r) { break; } return <warning descr="Require a positive PID from nonblocking wait.">$r !== -1</warning>; }
function direct($pid) { $x = 0; if ($x === 0) {} return <warning descr="Require a positive PID from nonblocking wait.">pcntl_waitpid($pid, $s, WNOHANG) !== -1</warning>; }
