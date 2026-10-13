<?php
// The compared variable no longer holds the raw wait status.
function incremented($child) { pcntl_waitpid($child, $status); $status++; if ($status === 3) { echo 1; } }
function decremented($child) { pcntl_waitpid($child, $status); --$status; if ($status === 3) { echo 1; } }
function destructured($child, $r) { pcntl_waitpid($child, $status); [$pid, $status] = $r; if ($status === 3) { echo 1; } }
function keyed($child, $r) { pcntl_waitpid($child, $status); list('s' => $status) = $r; if ($status === 3) { echo 1; } }
function looped($child, $r) { pcntl_waitpid($child, $status); foreach ($r as $status) {} if ($status === 3) { echo 1; } }
function removed($child) { pcntl_wait($status); unset($status); if ($status === 3) { echo 1; } }
function reassigned($child) { pcntl_waitpid($child, $status); $status = pcntl_wexitstatus($status); if ($status === 3) { echo 1; } }
function shifted($child) { pcntl_waitpid($child, $status); $status >>= 8; if ($status === 3) { echo 1; } }
function element($child) { pcntl_waitpid($child, $status); $status['code'] = 3; if ($status === 3) { echo 1; } }
function aliased($child) { pcntl_waitpid($child, $status); $ref = &$status; $ref = 3; if ($status === 3) { echo 1; } }
function captured($child) { pcntl_waitpid($child, $status); $f = function () use (&$status) { $status = 3; }; $f(); if ($status === 3) { echo 1; } }
function imported($child) { pcntl_waitpid($child, $status); global $status; if ($status === 3) { echo 1; } }
function repeated($child, $n) { pcntl_waitpid($child, $status); while ($n--) { if ($status === 3) { echo 1; } $status = $status >> 8; } }
