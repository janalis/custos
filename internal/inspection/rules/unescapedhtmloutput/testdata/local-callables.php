<?php
function rawClosureOutput() {
    $render = fn($text) => $text;
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$render($_GET['text'])</error> . '</p>';
}
function escapedClosureOutput() {
    $render = fn($text) => htmlspecialchars($text);
    echo '<p>' . $render($_GET['text']) . '</p>';
}
function capturedClosureOutput() {
    $text = $_GET['text'];
    $render = function () use ($text) { return $text; };
    $text = 'local';
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$render()</error> . '</p>';
}
class RendererOne { final public function render($text) { return $text; } }
class RendererTwo { final public function render($text) { return $text; } }
function finalMethodOutput(RendererOne|RendererTwo $renderer) {
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$renderer->render($_GET['text'])</error> . '</p>';
}
final class ReferenceWriter {
    public function safe(&$output) { $output = ['text' => 'local']; }
    public function retain(&$output) {}
    public static function safeStatic(&$output) { $output = ['text' => 'local']; }
}
function methodReferenceOutput(ReferenceWriter $writer) {
    $writer->safe($_GET);
    echo '<p>' . $_GET['text'] . '</p>';
}
function staticReferenceOutput() {
    ReferenceWriter::safeStatic($_GET);
    echo '<p>' . $_GET['text'] . '</p>';
}
function retainedMethodReferenceOutput(ReferenceWriter $writer) {
    $writer->retain($_GET);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$_GET['text']</error> . '</p>';
}
