<?php
class ReplaceStream
{
    public function __construct(&$stream)
    {
        $stream = fopen('replacement.data', 'r');
    }
}
$stream = fopen('original.data', 'r');
fclose($stream);
new ReplaceStream($stream);
fread($stream, 1);
