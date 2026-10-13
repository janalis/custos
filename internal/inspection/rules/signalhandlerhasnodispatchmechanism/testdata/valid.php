<?php
pcntl_async_signals(false); pcntl_signal(SIGTERM,function(){echo "stop";}); while (true) { pcntl_signal_dispatch(); usleep(10000); }
