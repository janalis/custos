<?php
declare(strict_types=1);

function strictTypesOnly() {
    pcntl_async_signals(false);
    pcntl_signal(SIGTERM, function () { echo "term"; });
    <warning descr="Dispatch signals when asynchronous handling is disabled.">while (true) {
        sleep(1);
    }</warning>
}
declare(ticks=1) {
    echo "ticking";
}
function afterTickBlock() {
    pcntl_async_signals(false);
    pcntl_signal(SIGTERM, function () { echo "term"; });
    <warning descr="Dispatch signals when asynchronous handling is disabled.">while (true) {
        sleep(1);
    }</warning>
}
declare(TICKS=1);
function afterTickStatement() {
    pcntl_async_signals(false);
    pcntl_signal(SIGTERM, function () { echo "term"; });
    while (true) {
        sleep(1);
    }
}
declare(ticks=0);
function ticksSwitchedOff() {
    pcntl_async_signals(false);
    pcntl_signal(SIGTERM, function () { echo "term"; });
    <warning descr="Dispatch signals when asynchronous handling is disabled.">while (true) {
        sleep(1);
    }</warning>
}
function tickBlockAround() {
    pcntl_async_signals(false);
    pcntl_signal(SIGTERM, function () { echo "term"; });
    declare(ticks=1) {
        while (true) {
            sleep(1);
        }
    }
}
