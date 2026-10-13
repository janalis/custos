<?php
if($signal=pcntl_sigwaitinfo([SIGTERM],$info)){handleSignal($signal);}

$broken = ;
