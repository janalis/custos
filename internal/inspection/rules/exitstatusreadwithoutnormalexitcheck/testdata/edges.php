<?php
pcntl_wexitstatus($unknown);$status=3;pcntl_wexitstatus($status);
function guardedStatus($pid){$r=pcntl_waitpid($pid,$s);if($r!==-1 && pcntl_wifexited($s)){pcntl_wexitstatus($s);}}
