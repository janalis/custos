<?php
function isType(string $file, string $ext): array {
    return [
        str_ends_with($file, $ext),
        str_ends_with($file, '.gz'),
        !str_ends_with(basename($file), $ext),
    ];
}
