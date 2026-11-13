<?php
// Before PHP 7.4 an element of null reads as null, silently.
class Item { public function price() { return 1; } }

function sum(?array $rows, ?Item $item)
{
    echo $rows['a'];
    return <warning descr="Check this value for null before the access: it is Item|null.">$item</warning>->price();
}
