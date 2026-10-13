<?php
declare(ticks=1);

// Statement-form ticks dispatch pending signals between statements.
pcntl_async_signals(false);
pcntl_signal(SIGTERM, function () { echo "term"; });
while (true) {
    usleep(1000);
}

function viaMethod($worker) {
    pcntl_async_signals(false);
    pcntl_signal(SIGTERM, function () { echo "term"; });
    while (true) {
        $worker->tick();
        sleep(1);
    }
}
