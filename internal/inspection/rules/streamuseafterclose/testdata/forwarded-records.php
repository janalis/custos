<?php
function passStreamRecord($record) { return $record; }
function forwardedRecordCaller() {
    $record = passStreamRecord(['stream' => tmpfile()]);
    fclose($record['stream']);
    <error descr="Use an open stream for this operation.">fread($record['stream'], 1)</error>;
}
function capturedRecordCaller() {
    $record = ['stream' => tmpfile()];
    $copy = fn() => $record;
    $alias = $copy();
    fclose($alias['stream']);
    <error descr="Use an open stream for this operation.">fread($record['stream'], 1)</error>;
}
function streamClosureFactory() { return fn() => ['stream' => tmpfile()]; }
function returnedClosureRecordCaller() {
    $factory = streamClosureFactory();
    $first = $factory();
    $second = $factory();
    fclose($first['stream']);
    <error descr="Use an open stream for this operation.">fread($first['stream'], 1)</error>;
    fread($second['stream'], 1);
}
function forwardedBorrowedParameter($stream) {
    fclose($stream);
    $record = passStreamRecord(['stream' => $stream]);
    <error descr="Use an open stream for this operation.">fread($record['stream'], 1)</error>;
}
