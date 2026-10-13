<?php
function open_named($path, $ctx, array $rest)
{
    $a = fopen(mode: 'rb', filename: 'php://memory');
    $b = fopen(mode: <warning descr="Add the 'b' flag to the mode for binary-safe file access.">'r'</warning>, filename: 'php://memory');
    $c = fopen($path, mode: <warning descr="Use the 'b' flag instead of 't' for binary-safe file access.">'wt'</warning>);
    $d = fopen(filename: $path, context: $ctx);
    $e = fopen($path, ...$rest);
    $f = fopen(...$rest);
    $g = fopen($path, "r\x62");
    $h = fopen($path, 'r\\');
    $quoted = "r'";
    $i = fopen($path, $quoted);
}
