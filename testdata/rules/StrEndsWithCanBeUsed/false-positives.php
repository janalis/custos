<?php
function isType(string $file, string $ext): array {
    return [
        substr($file, -strlen($ext)) === '.txt',
        substr($file, -strlen($ext), 2) === $ext,
        substr($file, -3) === $ext,
        substr($file, -(strlen($ext))) === $ext,
        mb_substr($file, -mb_strlen($ext, 'UTF-8')) === $ext,
        substr($file, -strlen($ext)) == $ext,
        (substr($file, -strlen($ext))) === $ext,
        substr($file, 3) === $ext,
    ];
}
