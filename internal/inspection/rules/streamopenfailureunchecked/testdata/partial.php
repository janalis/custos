<?php
// A partial application stores a closure: no failed stream reaches the write.
function openedLater($path, $condition)
{
    $stream = "fallback";
    if ($condition) {
        $stream = fopen($path, "r", ...);
    }
    fwrite($stream, "text");
}
