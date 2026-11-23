<?php
// A partial application holds a closure, not a stream opened for reading.
function writeThroughClosure(string $path): void
{
    $open = fopen(?, 'r');
    fwrite($open, 'x');
    $handle = fopen($path, 'r', ...);
    fwrite($handle, 'x');
}
