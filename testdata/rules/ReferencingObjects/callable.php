<?php
function invoke(callable &$handler, ?callable &$fallback, \Callable &$other) {}
function wrap(<warning descr="Objects are handed over by handle already; drop the '&' before '$handler'.">\Closure &$handler</warning>) {}
function hold(<warning descr="Objects are handed over by handle already; drop the '&' before '$target'.">object &$target</warning>) {}
function reset_target(object &$target) {
    $target = new stdClass();
}
