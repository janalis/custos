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
        <weak_warning descr="Call it directly: 'is_dir($dir)'.">call_user_func('is_dir', $dir)</weak_warning>,
        call_user_func('FileSize', $dir),
        call_user_func('fileatime', $dir),
        call_user_func('is_file', $dir),
        <weak_warning descr="Call it directly: 'is_link($dir)'.">call_user_func('is_link', $dir)</weak_warning>,
        <weak_warning descr="Call it directly: '\is_readable($dir)'.">call_user_func('\is_readable', $dir)</weak_warning>,
    ];
}
