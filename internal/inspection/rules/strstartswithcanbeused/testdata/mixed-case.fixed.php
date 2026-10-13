<?php

function routed(string $path): array
{
    return [
        str_starts_with($path, '/v2'),
        0 !== Mb_StrPos($path, '#'),
    ];
}
