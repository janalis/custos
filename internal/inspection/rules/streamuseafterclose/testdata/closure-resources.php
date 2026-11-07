<?php
function streamFactory() {
    $open = fn() => fopen('input.txt', 'rb');
    $h = $open();
    if ($h === false) { return; }
    fclose($h);
    <error descr="Use an open stream for this operation.">fread($h, 1)</error>;
}
function independentClosureResources() {
    $open = fn() => tmpfile();
    $first = $open();
    $second = $open();
    if ($first === false || $second === false) { return; }
    fclose($first);
    fread($second, 1);
}
function capturedClosedResource() {
    $h = fopen('input.txt', 'rb');
    if ($h === false) { return; }
    $peek = fn() => $h;
    fclose($h);
    $alias = $peek();
    <error descr="Use an open stream for this operation.">fread($alias, 1)</error>;
}
