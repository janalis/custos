<?php
namespace Shop;

final class Order {}

// With flow.assertions, assert() is a guard for every target PHP version.
function asserted(?Order $order): void
{
    assert($order !== null);
    if (<warning descr="This check is always true: the value is Shop\Order.">$order instanceof Order</warning>) {
        echo 1;
    }
}
