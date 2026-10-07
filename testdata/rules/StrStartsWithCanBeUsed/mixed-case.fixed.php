<?php

function routed(string $path): array
{
    return [
        str_starts_with($path, '/v2'),
        !str_starts_with($path, '#'),
    ];
}
