<?php
function unpack(string $file): void
{
    $zip = new ZipArchive();
    if (!<warning descr="Accept only true from ZipArchive::open.">$zip->open($file)</warning>) {
        throw new RuntimeException('Cannot open.');
    }
    while (<warning descr="Accept only true from ZipArchive::open.">$zip->open($file)</warning>) {
        break;
    }
}
