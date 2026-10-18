<?php
// Any rebinding between the wait and the comparison leaves no raw status.
function unsetFirst($child) { pcntl_waitpid($child, $status); unset($status); if ($status === 3) { echo 1; } }
function destructured($child) { pcntl_waitpid($child, $status); [$status] = [3]; if ($status === 3) { echo 1; } }
function incremented($child) { pcntl_waitpid($child, $status); $status++; if ($status === 3) { echo 1; } }
function iterated($child, $codes) { pcntl_waitpid($child, $status); foreach ($codes as $status) { } if ($status === 3) { echo 1; } }
function captured($child) { pcntl_waitpid($child, $status); $reset = function () use (&$status) { $status = 0; }; $reset(); if ($status === 3) { echo 1; } }
function capturedLater($child, $more) { pcntl_waitpid($child, $status); while ($more) { if ($status === 3) { echo 1; } $reset = function () use (&$status) { $status = 0; }; $reset(); } }
