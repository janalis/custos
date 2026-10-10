<?php
// @custos-ignore WaitStatusUsedAsExitCode
$pid=pcntl_waitpid($child,$status); if ($status===3) { echo "exit three"; }
