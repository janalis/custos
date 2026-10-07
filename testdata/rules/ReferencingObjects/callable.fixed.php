<?php
function invoke(callable &$handler, ?callable &$fallback, \Callable &$other) {}
function wrap(\Closure $handler) {}
function hold(object $target) {}
function reset_target(object &$target) {
    $target = new stdClass();
}
