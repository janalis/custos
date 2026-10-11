<?php
// Reads, by-value captures and later writes leave the raw status in place.
function copied($child) { pcntl_waitpid($child, $status); $code = $status; if (<warning descr="Decode the child wait status.">$status === 3</warning>) { echo 1; } }
function byValue($child) { pcntl_waitpid($child, $status); $f = function () use ($status) { return $status; }; if (<warning descr="Decode the child wait status.">$status === 3</warning>) { echo 1; } }
function written($child) { pcntl_wait($status); if (<warning descr="Decode the child wait status.">$status > 1</warning>) { echo 1; } $status = 0; }
function polled($child) { for (;;) { pcntl_waitpid($child, $status); if (<warning descr="Decode the child wait status.">$status === 1</warning>) { echo 1; } $status = 0; } }
function other($child, $r) { pcntl_waitpid($child, $status); foreach ($r as $item) { $ref = &$item; if (<warning descr="Decode the child wait status.">$status === 3</warning>) { echo 1; } } }
function before($child) { $status = 0; $status++; pcntl_waitpid($child, $status); if (<warning descr="Decode the child wait status.">$status === 3</warning>) { echo 1; } }
