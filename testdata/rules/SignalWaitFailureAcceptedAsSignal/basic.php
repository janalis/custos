<?php
if(<warning descr="Reject signal-wait failure before handling the result.">$signal=pcntl_sigwaitinfo([SIGTERM],$info)</warning>){handleSignal($signal);}
