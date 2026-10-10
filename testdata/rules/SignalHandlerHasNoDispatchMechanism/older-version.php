<?php
pcntl_async_signals(false); pcntl_signal(SIGTERM,function(){echo "stop";}); while (true) { usleep(10000); }
