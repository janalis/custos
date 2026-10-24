<?php
function open_named($path, $ctx, array $rest)
{
    $a = fopen(mode: 'rb', filename: 'data.bin');
    $b = fopen(mode: 'rb', filename: 'data.bin');
    $c = fopen($path, mode: 'wb');
    $d = fopen(filename: $path, context: $ctx);
    $e = fopen($path, ...$rest);
    $f = fopen(...$rest);
    $g = fopen($path, "r\x62");
    $h = fopen($path, 'r\\');
    $quoted = "r'";
    $i = fopen($path, $quoted);
}
