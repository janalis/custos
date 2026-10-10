<?php
// @noinspection ForkChildFallsIntoParentWork
$pid=pcntl_fork(); if ($pid===0) { echo "child"; } echo "continuation";
