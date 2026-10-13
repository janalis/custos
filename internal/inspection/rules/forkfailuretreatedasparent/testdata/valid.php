<?php
$pid=pcntl_fork(); if ($pid===-1) { throw new RuntimeException(); } if ($pid>0) { echo "parent"; } else { echo "child"; }
