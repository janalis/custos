<?php
namespace Probe;

use function Other\is_file;
use function is_link;

final class Fake { public static ?float $free = null; }

function disk_free_space(string $dir): float|false
{
    $real = call_user_func('disk_free_space', $dir);
    return Fake::$free ?? $real;
}

function DISK_TOTAL_SPACE(string $dir): float|false
{
    return call_user_func('\disk_total_space', $dir) ?: 0.0;
}

function sample(string $dir)
{
    return [
        is_dir($dir),
        call_user_func('Disk_Free_Space', $dir),
        call_user_func('disk_total_space', $dir),
        call_user_func('is_file', $dir),
        is_link($dir),
        \is_readable($dir),
    ];
}
