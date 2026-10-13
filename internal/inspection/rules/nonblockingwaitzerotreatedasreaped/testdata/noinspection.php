<?php
// @noinspection NonblockingWaitZeroTreatedAsReaped
$r=pcntl_waitpid($pid,$status,WNOHANG); if ($r!==-1) { echo pcntl_wexitstatus($status); }
