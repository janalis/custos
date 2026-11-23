<?php
// A partial application holds a closure, not a stream opened for writing.
function readThroughClosure(string $path): void
{
    $open = fopen(?, 'w');
    fread($open, 10);
    $handle = fopen($path, 'w', ...);
    fread($handle, 10);
}
