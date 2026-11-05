<?php
function rawClosureFactory() { return fn($text) => $text; }
function safeClosureFactory(string $prefix) { return fn($text) => htmlspecialchars($prefix . $text); }
final class RawOutputObject { public function __invoke($text) { return $text; } }
final class SafeOutputObject { public function __invoke($text) { return htmlspecialchars($text); } }
function extendedCallbacks($flag) {
    $raw = rawClosureFactory();
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$raw($_GET['text'])</error>;
    $safe = safeClosureFactory('prefix');
    echo '<p>' . $safe($_GET['text']);
    $object = new RawOutputObject();
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$object($_GET['text'])</error>;
    $escaped = new SafeOutputObject();
    echo '<p>' . $escaped($_GET['text']);
    $choice = $flag ? 'strtolower' : 'strtoupper';
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$choice($_GET['text'])</error>;
    $mapped = array_map('htmlspecialchars', [$_GET['text']]);
    echo '<p>' . $mapped[0];
    $mappedRaw = array_map(fn($text) => $text, [$_GET['text']]);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$mappedRaw[0]</error>;
}
