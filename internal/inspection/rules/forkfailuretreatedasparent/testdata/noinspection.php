<?php
// @noinspection ForkFailureTreatedAsParent
$pid=pcntl_fork(); if ($pid) { echo "parent"; } else { echo "child"; }
