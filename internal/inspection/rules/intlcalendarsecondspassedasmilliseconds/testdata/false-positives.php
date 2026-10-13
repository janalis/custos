<?php
function referenceCapture(IntlCalendar $c) {
    $t = time();
    $toMillis = function () use (&$t) { $t *= 1000; };
    $toMillis();
    $c->setTime($t);
}
function catchBinding(IntlCalendar $c) {
    $t = time();
    try { load(); } catch (Exception $t) { }
    $c->setTime($t);
}
