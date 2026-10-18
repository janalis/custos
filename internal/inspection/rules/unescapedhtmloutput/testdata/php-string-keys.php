<?php
function slashLetterKey() {
    $values = ["\a" => $_GET['value'], "\x07" => 'safe'];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$values["\a"]</error>;
}
function distinctControlKey() {
    $values = ["\a" => 'safe', "\x07" => $_GET['value']];
    echo '<p>' . $values["\a"];
}
function slashUnicodeKey() {
    $values = ["\u0030" => $_GET['value'], 0 => 'safe'];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$values["\u0030"]</error>;
}
function distinctNumericKey() {
    $values = ["\u0030" => 'safe', 0 => $_GET['value']];
    echo '<p>' . $values["\u0030"];
}
function binaryRequestName() {
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">${b"_GET"}['value']</error>;
}
