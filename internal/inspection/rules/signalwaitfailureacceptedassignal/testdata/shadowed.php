<?php namespace Independent;
function pcntl_sigwaitinfo(){}

if($signal=pcntl_sigwaitinfo([SIGTERM],$info)){handleSignal($signal);}
