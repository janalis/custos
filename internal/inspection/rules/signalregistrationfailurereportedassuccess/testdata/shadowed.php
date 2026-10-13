<?php namespace Independent;
function pcntl_signal(){}

function install($h){pcntl_signal(SIGTERM,$h);return true;}
