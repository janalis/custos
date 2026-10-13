<?php
pcntl_async_signals(false); pcntl_signal(SIGTERM,function(){echo "stop";}); <warning descr="Dispatch signals when asynchronous handling is disabled.">while (true) { usleep(10000); }</warning>
