<?php
function isType(string $file, string $ext): array {
    return [
        substr($file, -strlen($ext)) === $ext,
        str_ends_with($file, '.gz'),
        substr(basename($file), -mb_strlen($ext)) !== $ext,
    ];
}
