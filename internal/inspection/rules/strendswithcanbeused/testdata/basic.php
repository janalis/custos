<?php
function isType(string $file, string $ext): array {
    return [
        <weak_warning descr="Replace with 'str_ends_with($file, $ext)'.">substr($file, -strlen($ext)) === $ext</weak_warning>,
        '.gz' === mb_substr($file, - mb_strlen('.gz')),
        substr(basename($file), -mb_strlen($ext)) !== $ext,
    ];
}
