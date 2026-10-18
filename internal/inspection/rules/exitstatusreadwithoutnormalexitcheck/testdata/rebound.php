<?php
// Any rebinding between the wait and the read leaves no raw status.
function unsetFirst($child) { pcntl_waitpid($child, $status); unset($status); echo pcntl_wexitstatus($status); }
function incremented($child) { pcntl_waitpid($child, $status); $status++; echo pcntl_wexitstatus($status); }
function destructured($child) { pcntl_waitpid($child, $status); [$status] = [0]; echo pcntl_wexitstatus($status); }
function iterated($child, $codes) { pcntl_waitpid($child, $status); foreach ($codes as $status) { } echo pcntl_wexitstatus($status); }
function captured($child) { pcntl_waitpid($child, $status); $reset = function () use (&$status) { $status = 0; }; $reset(); echo pcntl_wexitstatus($status); }
function element($child) { pcntl_waitpid($child, $status); $status[0] = 1; echo pcntl_wexitstatus($status); }
