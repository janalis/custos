<?php
function install($h){pcntl_signal(SIGTERM,$h);return true;}

$broken = ;
