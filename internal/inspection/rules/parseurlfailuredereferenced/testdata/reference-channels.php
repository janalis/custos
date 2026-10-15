<?php
final class ChannelWriter {
    public function replace(&$parts, string $url) { $parts = parse_url($url); }
}
function reference_capture(string $url, string $replacement, $n) {
    $parts = parse_url($url);
    if ($parts === false) { return; }
    $write = function () use (&$parts, $replacement) { $parts = parse_url($replacement); };
again:
    echo <warning descr="Reject URL parsing failure before indexing.">$parts['host']</warning>;
    $write();
    if ($n--) { goto again; }
}
function reference_method(ChannelWriter $writer, string $url, string $replacement, $n) {
    $parts = parse_url($url);
    if ($parts === false) { return; }
    while ($n--) {
        echo <warning descr="Reject URL parsing failure before indexing.">$parts['host']</warning>;
        $writer->replace($parts, $replacement);
    }
}
function global_replace(string $url) { global $parts; $parts = parse_url($url); }
function global_channel(string $url, string $replacement, $n) {
    global $parts;
    $parts = parse_url($url);
    if ($parts === false) { return; }
    while ($n--) {
        echo <warning descr="Reject URL parsing failure before indexing.">$parts['host']</warning>;
        global_replace($replacement);
    }
}
function reference_element($service, string $url) {
    $parts = parse_url($url);
    if ($parts === false) { return; }
    $service->rewrite($parts['host']);
    echo $parts['path'];
}
function checked_replacement(ChannelWriter $writer, string $url, $n) {
    $parts = parse_url($url);
    while ($n--) {
        if ($parts === false) { return; }
        echo $parts['host'];
        $writer->replace($parts, $url);
    }
}
