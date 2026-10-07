<?php
function encode($data, bool $strict)
{
    $flags = 0;
    if ($strict) {
        $flags |= JSON_THROW_ON_ERROR;
    }
    $opts = JSON_PRETTY_PRINT;
    $opts += JSON_UNESCAPED_SLASHES;
    return [json_encode($data, $flags), json_decode($data, true, 512, $opts)];
}
