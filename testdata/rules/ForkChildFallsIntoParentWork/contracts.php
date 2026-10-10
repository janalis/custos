<?php
$pid=pcntl_fork(); if (<warning descr="Terminate or separate the fork child branch.">$pid===0</warning>) { echo "child"; } echo "continuation";
