<?php
function has_city(array $order, $customer)
{
    $a = isset(<weak_warning descr="Redundant isset() argument: a deeper array access already covers it.">$order</weak_warning>, $order['ship']['city']);
    $b = isset($order['ship']['city'], <weak_warning descr="Redundant isset() argument: a deeper array access already covers it.">$order['ship']</weak_warning>);
    $c = isset(
        <weak_warning descr="Redundant isset() argument: a deeper array access already covers it.">$customer->tags</weak_warning>,
        $customer->tags[0],
        $customer
    );
    $g = isset(
        <weak_warning descr="Redundant isset() argument: a deeper array access already covers it.">$order</weak_warning>,
        $order['a']['b'],
        <weak_warning descr="Redundant isset() argument: a deeper array access already covers it.">$order['a']</weak_warning>
    );
    $k = isset($order['a']['b'], <weak_warning descr="Redundant isset() argument: a deeper array access already covers it.">$order</weak_warning>, <weak_warning descr="Redundant isset() argument: a deeper array access already covers it.">$order['a']</weak_warning>);
    return [$a, $b, $c, $g, $k];
}
