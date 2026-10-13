<?php

function archived(string $name, string $suffix): array
{
    return [
        SubStr($name, -StrLen($suffix)) === $suffix,
        '.tar' === MB_SUBSTR($name, -MB_STRLEN('.tar')),
    ];
}
