<?php
// @noinspection SignalWaitFailureAcceptedAsSignal

if($signal=pcntl_sigwaitinfo([SIGTERM],$info)){handleSignal($signal);}
