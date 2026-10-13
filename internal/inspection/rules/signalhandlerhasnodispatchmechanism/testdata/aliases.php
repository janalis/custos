<?php
use function pcntl_async_signals as builtinCall12;
builtinCall12(false); pcntl_signal(SIGTERM,function(){echo "stop";}); <warning descr="Dispatch signals when asynchronous handling is disabled.">while (true) { usleep(10000); }</warning>
