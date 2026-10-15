<?php
function callback_filter_reentry(string $url, array $filters) {
    static $parts;
    $parts = parse_url($url);
    if ($parts === false) { return; }
    filter_input_array(INPUT_POST, ['field' => ['filter' => FILTER_CALLBACK, 'options' => 'hook']]);
    echo <warning descr="Reject URL parsing failure before indexing.">$parts['host']</warning>;
}
function constant_filter_control(string $url) {
    static $parts;
    $parts = parse_url($url);
    if ($parts === false) { return; }
    filter_input_array(INPUT_POST, FILTER_VALIDATE_INT, false);
    echo $parts['host'];
}
function class_autoload_reentry(string $url, string $class, SplFileInfo $file) {
    static $parts;
    $parts = parse_url($url);
    if ($parts === false) { return; }
    class_alias($class, 'GeneratedAlias');
    echo <warning descr="Reject URL parsing failure before indexing.">$parts['host']</warning>;
}
function file_info_class_reentry(string $url, string $class, SplFileInfo $file) {
    static $parts;
    $parts = parse_url($url);
    if ($parts === false) { return; }
    $file->getFileInfo($class);
    echo <warning descr="Reject URL parsing failure before indexing.">$parts['host']</warning>;
}
function warning_callback_reentry(string $url, string $bytes) {
    static $parts;
    $parts = parse_url($url);
    if ($parts === false) { return; }
    call_user_func('hex2bin', $bytes);
    echo <warning descr="Reject URL parsing failure before indexing.">$parts['host']</warning>;
}
function scalar_callback_control(string $url, string $bytes) {
    static $parts;
    $parts = parse_url($url);
    if ($parts === false) { return; }
    call_user_func('strlen', $bytes);
    echo $parts['host'];
}
