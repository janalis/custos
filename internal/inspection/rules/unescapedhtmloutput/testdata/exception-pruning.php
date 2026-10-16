<?php
function inertExpression() {
    $text = 'ready';
    try { 1; 'fixed'; true; }
    catch (Throwable $error) { $text = $_GET['text']; }
    echo '<p>' . $text;
}
function freshScalarWrite() {
    $text = 'ready';
    try { $count = 1; $count = 2; $count; }
    catch (Throwable $error) { $text = $_GET['text']; }
    echo '<p>' . $text;
}
function possibleCall() {
    $text = 'ready';
    try { work(); }
    catch (Throwable $error) { $text = $_GET['text']; }
    <error descr="Escape untrusted text for its HTML output context.">echo '<p>' . $text;</error>
}
class ReleasingValue { public function __destruct() { throw new Exception; } }
function discardedObject() {
    $text = 'ready';
    $value = new ReleasingValue;
    try { $value = 1; }
    catch (Throwable $error) { $text = $_GET['text']; }
    <error descr="Escape untrusted text for its HTML output context.">echo '<p>' . $text;</error>
}
