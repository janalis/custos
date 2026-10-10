<?php namespace Independent;
function pcntl_exec(){}

pcntl_exec("/missing/program",[]);exit(0);
