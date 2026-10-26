<?php
function unpack(string $file): void
{
    $zip = new ZipArchive();
    if (true !== $zip->open($file)) {
        throw new RuntimeException('Cannot open.');
    }
    while (true === $zip->open($file)) {
        break;
    }
}
