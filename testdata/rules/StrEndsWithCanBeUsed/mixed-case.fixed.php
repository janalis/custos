<?php

function archived(string $name, string $suffix): array
{
    return [
        str_ends_with($name, $suffix),
        str_ends_with($name, '.tar'),
    ];
}
