<?php
$pid=pcntl_waitpid($child,$status); if ($pid>0 && pcntl_wifexited($status) && pcntl_wexitstatus($status)===3) { echo "exit three"; }
