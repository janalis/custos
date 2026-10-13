<?php
function resumedInSameStatement() {
    $g = (function () { yield 1; return 2; })();
    return [$g->next(), $g->getReturn()];
}
function resumedByEarlierArgument() {
    $g = (function () { yield 1; return 2; })();
    var_dump($g->next(), $g->getReturn());
}
function conditionalRead($skip) {
    $g = (function () { yield 1; return 2; })();
    return $skip ? null : $g->getReturn();
}
function deferredRead() {
    $g = (function () { yield 1; return 2; })();
    return fn () => $g->getReturn();
}
function finishedFirst() {
    $g = (function () { yield 1; return 2; })();
    foreach ($g as $v) {
    }
    return $g->getReturn();
}
