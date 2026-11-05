<?php
final class EngineRenderer {
    public static function raw($text) { return $text; }
    public function safe($text) { return htmlspecialchars($text); }
}
function engineCallables($flag) {
    $native = strtolower(...);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$native($_GET['text'])</error>;
    echo '<p>' . call_user_func('htmlspecialchars', $_GET['text']);
    $static = EngineRenderer::raw(...);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$static($_GET['text'])</error>;
    $renderer = new EngineRenderer();
    echo '<p>' . call_user_func([$renderer, 'safe'], $_GET['text']);
    $choice = $flag ? fn($text) => $text : fn($text) => $text;
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$choice($_GET['text'])</error>;
    $safe = $flag ? htmlspecialchars(...) : fn($text) => htmlspecialchars($text);
    echo '<p>' . $safe($_GET['text']);
}
