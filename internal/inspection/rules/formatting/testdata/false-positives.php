<?php

namespace App;

function total($a, $b)
{
    if ($a > $b) {
        return $a;
    }

    return $b;
}

// @custos-format-ignore-start
$matrix = [
    1,0,0,
    0,1,0,
    0,0,1,
];
// @custos-format-ignore-end

final class Cart
{
    private array $items = [];

    public function add(string $item): void
    {
        $this->items[] = $item;
    }
}
