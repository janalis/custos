<?php
function slashLetterKey() {
    $values = ["\a" => $_GET['value'], "\x07" => 'safe'];
    <error descr="Escape untrusted text for its HTML output context.">echo '<p>' . $values["\a"];</error>
}
function distinctControlKey() {
    $values = ["\a" => 'safe', "\x07" => $_GET['value']];
    echo '<p>' . $values["\a"];
}
function slashUnicodeKey() {
    $values = ["\u0030" => $_GET['value'], 0 => 'safe'];
    <error descr="Escape untrusted text for its HTML output context.">echo '<p>' . $values["\u0030"];</error>
}
function distinctNumericKey() {
    $values = ["\u0030" => 'safe', 0 => $_GET['value']];
    echo '<p>' . $values["\u0030"];
}
function binaryRequestName() {
    <error descr="Escape untrusted text for its HTML output context.">echo '<p>' . ${b"_GET"}['value'];</error>
}
