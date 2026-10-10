<?php
$pid=pcntl_fork(); if (<warning descr="Handle fork failure separately.">$pid</warning>) { echo "parent"; } else { echo "child"; }
