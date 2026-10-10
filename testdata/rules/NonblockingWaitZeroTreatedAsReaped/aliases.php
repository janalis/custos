<?php
use function pcntl_waitpid as builtinCall11;
$r=builtinCall11($pid,$status,WNOHANG); if (<warning descr="Require a positive PID from nonblocking wait.">$r!==-1</warning>) { echo pcntl_wexitstatus($status); }
