<?php
function isType(string $file, string $ext): array {
    return [
        <weak_warning descr="Replace with 'str_ends_with($file, $ext)'.">substr($file, -strlen($ext)) === $ext</weak_warning>,
        <weak_warning descr="Replace with 'str_ends_with($file, '.gz')'.">'.gz' === mb_substr($file, - mb_strlen('.gz'))</weak_warning>,
        <weak_warning descr="Replace with '!str_ends_with(basename($file), $ext)'.">substr(basename($file), -mb_strlen($ext)) !== $ext</weak_warning>,
    ];
}
