<?php
function matched($pid) { $r = pcntl_waitpid($pid, $s, WNOHANG); return match ($r) { 0 => 'running', default => $r !== -1 }; }
