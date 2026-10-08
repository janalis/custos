<?php
namespace Probe;

use function Other\is_file;
use function is_link;

final class Fake { public static ?int $size = null; }

function filesize(string $dir): int|false
{
    $real = call_user_func('filesize', $dir);
    return Fake::$size ?? $real;
}

function FILEATIME(string $dir): int|false
{
    return call_user_func('\fileatime', $dir) ?: 0;
}

function sample(string $dir)
{
    return [
        is_dir($dir),
        call_user_func('FileSize', $dir),
        call_user_func('fileatime', $dir),
        call_user_func('is_file', $dir),
        is_link($dir),
        \is_readable($dir),
    ];
}
