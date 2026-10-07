<?php
function has_city(array $order, $customer)
{
    $a = isset($order['ship']['city']);
    $b = isset($order['ship']['city']);
    $c = isset(
        $customer->tags[0],
        $customer
    );
    $g = isset(
        $order['a']['b']
    );
    $k = isset($order['a']['b']);
    return [$a, $b, $c, $g, $k];
}
