<?php
function isType(string $file, string $ext): array {
    return [
        substr($file, -strlen($ext)) === $ext,
        '.gz' === mb_substr($file, - mb_strlen('.gz')),
        substr(basename($file), -mb_strlen($ext)) !== $ext,
    ];
}
