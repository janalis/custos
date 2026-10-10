<?php
$r=pcntl_waitpid($pid,$status,WNOHANG); if ($r>0 && pcntl_wifexited($status)) { echo pcntl_wexitstatus($status); }
