<?php

function archived(string $name, string $suffix): array
{
    return [
        SubStr($name, -StrLen($suffix)) === $suffix,
        str_ends_with($name, '.tar'),
    ];
}
