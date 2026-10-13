<?php
function inCondition() {
    pcntl_async_signals(false);
    pcntl_signal(SIGTERM, function () { echo "term"; });
    while (pcntl_signal_dispatch()) {
        sleep(1);
    }
}
function methodInCondition($loop) {
    pcntl_async_signals(false);
    pcntl_signal(SIGTERM, function () { echo "term"; });
    while ($loop->tick()) {
        usleep(10);
    }
}
function viaMethod($worker) {
    pcntl_async_signals(false);
    pcntl_signal(SIGTERM, function () { echo "term"; });
    while (true) {
        $worker?->tick();
        sleep(1);
    }
}
function viaStatic() {
    pcntl_async_signals(false);
    pcntl_signal(SIGTERM, function () { echo "term"; });
    while (true) {
        Signals::dispatch();
        usleep(100);
    }
}
function viaInclude() {
    pcntl_async_signals(false);
    pcntl_signal(SIGTERM, function () { echo "term"; });
    while (true) {
        include 'tick.php';
        sleep(1);
    }
}
function viaConstructor() {
    pcntl_async_signals(false);
    pcntl_signal(SIGTERM, function () { echo "term"; });
    while (true) {
        new Ticker();
        sleep(1);
    }
}
function viaEval($code) {
    pcntl_async_signals(false);
    pcntl_signal(SIGTERM, function () { echo "term"; });
    while (true) {
        eval($code);
        sleep(1);
    }
}
function viaYield() {
    pcntl_async_signals(false);
    pcntl_signal(SIGTERM, function () { echo "term"; });
    while (true) {
        yield 1;
        sleep(1);
    }
}
function viaDelegation($inner) {
    pcntl_async_signals(false);
    pcntl_signal(SIGTERM, function () { echo "term"; });
    while (true) {
        yield from $inner;
        sleep(1);
    }
}
function viaDynamicCall($tick) {
    pcntl_async_signals(false);
    pcntl_signal(SIGTERM, function () { echo "term"; });
    while (true) {
        $tick();
        sleep(1);
    }
}
function onlySleeps() {
    pcntl_async_signals(false);
    pcntl_signal(SIGTERM, function () { echo "term"; });
    <warning descr="Dispatch signals when asynchronous handling is disabled.">while (true) {
        sleep(1);
    }</warning>
}
