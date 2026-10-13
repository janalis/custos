<?php
// @custos-ignore ForkFailureTreatedAsParent
$pid=pcntl_fork(); if ($pid) { echo "parent"; } else { echo "child"; }
