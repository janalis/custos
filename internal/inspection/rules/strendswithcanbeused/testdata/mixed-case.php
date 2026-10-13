<?php

function archived(string $name, string $suffix): array
{
    return [
        <weak_warning descr="Replace with 'str_ends_with($name, $suffix)'.">SubStr($name, -StrLen($suffix)) === $suffix</weak_warning>,
        '.tar' === MB_SUBSTR($name, -MB_STRLEN('.tar')),
    ];
}
