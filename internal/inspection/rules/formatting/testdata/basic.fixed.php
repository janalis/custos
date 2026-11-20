<?php

namespace App;

function total($a, $b)
{
    if ($a > $b) {
        return $a;
    }
    return $b;
}

class Cart
{
    private $items = [];
    public function add($item)
    {
        $this->items[] = $item;
    }
}
