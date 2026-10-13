<?php
$pid=pcntl_fork(); if ($pid===0) { echo "child"; exit; } echo "continuation";
