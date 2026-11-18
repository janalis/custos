<?php
$r=pcntl_waitpid($pid,$status,WNOHANG); if (<warning descr="Require a positive PID from nonblocking wait.">$r!==-1</warning>) { echo pcntl_wexitstatus($status); }
