<?php
namespace Store;

function normalize($qty, $ratio, $label, $item, $price) {
    $a = (int)$qty;
    $b = (float)($ratio * 2);
    $c = (bool)($qty ?: $ratio);
    $d = (int)$label;
    $e = intval($label, 8);
    $f = (string)$price;
    $first = ((string)$price)[0];

    $qty = (int)$qty;
    $item['tags'] = (array)$item['tags'];
    $ratio = (float)$ratio;
    settype($ratio, 'null');
    $ok = settype($ratio, 'int');

    $g = (string)$label;
    $h = (string)($item->name);
    $i = "#$label";
    $j = "{$label}{$qty}";
    $k = (string)$price;
    return compact('a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'ok', 'first');
}

class Tag extends Base {
    public function __toString()
    {
        return 'tag:' . parent::__toString();
    }
}
