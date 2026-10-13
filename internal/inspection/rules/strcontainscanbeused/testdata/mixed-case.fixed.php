<?php

function scan(string $row, string $mark): array
{
    return [
        str_contains($row, $mark),
        MB_STRPOS($row, '|') === FALSE,
    ];
}
