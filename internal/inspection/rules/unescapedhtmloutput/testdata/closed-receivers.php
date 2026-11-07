<?php
final class RawRendererA { public function render($text) { return $text; } }
final class RawRendererB { public function render($text) { return $text; } }
final class SafeRendererA { public function render($text) { return htmlspecialchars($text); } }
final class SafeRendererB { public function render($text) { return htmlspecialchars($text); } }
function closedRaw(RawRendererA|RawRendererB $renderer) {
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$renderer->render($_GET['text'])</error> . '</p>';
}
function closedSafe(SafeRendererA|SafeRendererB $renderer) {
    echo '<p>' . $renderer->render($_GET['text']) . '</p>';
}
function inferredRaw($flag) {
    $renderer = $flag ? new RawRendererA : new RawRendererB;
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$renderer->render($_GET['text'])</error> . '</p>';
}
