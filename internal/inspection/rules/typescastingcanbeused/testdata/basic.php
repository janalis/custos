<?php
namespace Store;

function normalize($qty, $ratio, $label, $item, $price) {
    $a = <weak_warning descr="Use '(int)$qty' instead (a cast is clearer and faster).">intval($qty)</weak_warning>;
    $b = <weak_warning descr="Use '(float)($ratio * 2)' instead (a cast is clearer and faster).">floatval($ratio * 2)</weak_warning>;
    $c = <weak_warning descr="Use '(bool)($qty ?: $ratio)' instead (a cast is clearer and faster).">\boolval($qty ?: $ratio)</weak_warning>;
    $d = <weak_warning descr="Use '(int)$label' instead (a cast is clearer and faster).">intval($label, 10)</weak_warning>;
    $e = intval($label, 8);
    $f = <weak_warning descr="Use '(string)$price' instead (a cast is clearer and faster).">strval($price)</weak_warning>;
    $first = <weak_warning descr="Use '((string)$price)' instead (a cast is clearer and faster).">strval($price)</weak_warning>[0];

    <weak_warning descr="Use '$qty = (int)$qty' instead (a cast is clearer and faster).">settype($qty, 'integer')</weak_warning>;
    <weak_warning descr="Use '$item['tags'] = (array)$item['tags']' instead (a cast is clearer and faster).">settype($item['tags'], "array")</weak_warning>;
    <weak_warning descr="Use '$ratio = (float)$ratio' instead (a cast is clearer and faster).">settype($ratio, 'double')</weak_warning>;
    settype($ratio, 'null');
    $ok = settype($ratio, 'int');

    $g = <weak_warning descr="Use '(string)$label' to make the string conversion explicit.">"$label"</weak_warning>;
    $h = <weak_warning descr="Use '(string)($item->name)' to make the string conversion explicit.">"{$item->name}"</weak_warning>;
    $i = "#$label";
    $j = "{$label}{$qty}";
    $k = <weak_warning descr="Use '(string)$price' instead of calling __toString() directly.">$price->__toString()</weak_warning>;
    return compact('a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'ok', 'first');
}

class Tag extends Base {
    public function __toString()
    {
        return 'tag:' . parent::__toString();
    }
}
