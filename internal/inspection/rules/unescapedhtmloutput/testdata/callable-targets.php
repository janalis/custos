<?php
function forwardText($text) { return $text; }
function escapeText($text) { return htmlspecialchars($text); }
function callableRaw() {
    $render = forwardText(...);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$render($_GET['text'])</error> . '</p>';
}
function callableSafe() {
    $render = escapeText(...);
    echo '<p>' . $render($_GET['text']) . '</p>';
    echo '<p>' . call_user_func('escapeText', $_GET['text']) . '</p>';
    echo '<p>' . call_user_func_array('escapeText', [$_GET['text']]) . '</p>';
}
function defaultArgumentRaw() {
    $render = fn($text = 'local') => $text;
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$render($_GET['text'])</error> . '</p>';
}
function variadicArgumentRaw() {
    $render = fn(...$texts) => $texts[0];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$render($_GET['text'])</error> . '</p>';
}
