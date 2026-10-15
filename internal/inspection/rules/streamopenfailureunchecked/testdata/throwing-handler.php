<?php
function handler_arrow($path, $condition) {
    set_error_handler(fn() => throw new RuntimeException());
    $stream = "fallback";
    if ($condition) { $stream = fopen($path, "r"); }
    fwrite($stream, "text");
}
function handler_closure($path, $condition) {
    set_error_handler(function() { throw new RuntimeException(); });
    $stream = "fallback";
    if ($condition) { $stream = fopen($path, "r"); }
    fwrite($stream, "text");
}
function handler_restored($path, $condition) {
    set_error_handler(fn() => throw new RuntimeException()); restore_error_handler();
    $stream = "fallback";
    if ($condition) { $stream = fopen($path, "r"); }
    <warning descr="Check that the stream opened successfully.">fwrite($stream, "text")</warning>;
}
function handler_replaced($path, $condition) {
    set_error_handler(fn() => throw new RuntimeException()); set_error_handler(fn() => false);
    $stream = "fallback";
    if ($condition) { $stream = fopen($path, "r"); }
    <warning descr="Check that the stream opened successfully.">fwrite($stream, "text")</warning>;
}
function handler_conditional($path, $condition) {
    if ($condition) { set_error_handler(fn() => throw new RuntimeException()); }
    $stream = "fallback";
    if ($condition) { $stream = fopen($path, "r"); }
    fwrite($stream, "text");
}
function handler_ordinary($path, $condition) {

    $stream = "fallback";
    if ($condition) { $stream = fopen($path, "r"); }
    <warning descr="Check that the stream opened successfully.">fwrite($stream, "text")</warning>;
}
function handler_nonthrowing($path, $condition) {
    set_error_handler(function() { return false; });
    $stream = "fallback";
    if ($condition) { $stream = fopen($path, "r"); }
    <warning descr="Check that the stream opened successfully.">fwrite($stream, "text")</warning>;
}
function mixed_uncertain_first($path,$condition) {
 if($condition){set_error_handler(fn() => throw new RuntimeException()); $stream=fopen($path,"r");}else{$stream=fopen($path,"r");}
 <warning descr="Check that the stream opened successfully.">fwrite($stream,"text")</warning>;
}
function mixed_ordinary_first($path,$condition) {
 if($condition){$stream=fopen($path,"r");}else{set_error_handler(fn() => throw new RuntimeException()); $stream=fopen($path,"r");}
 <warning descr="Check that the stream opened successfully.">fwrite($stream,"text")</warning>;
}
