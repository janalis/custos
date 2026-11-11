<?php
// No strict_types: PHP coerces scalars, but never arrays, objects or null.
namespace Shop;

class Order {}

function charge(int $cents, string $currency = 'EUR') {}
function receipt(Order $order) {}
function note(string $text = null) {}

charge('1200');
charge(12.0, 978);
charge(true);
charge(<error descr="Pass int to $cents of Shop\charge(), not array.">[]</error>);
charge(<error descr="Pass int to $cents of Shop\charge(), not null.">null</error>);
charge(<error descr="Pass int to $cents of Shop\charge(), not Shop\Order.">new Order()</error>);
receipt(<error descr="Pass Shop\Order to $order of Shop\receipt(), not string.">'order-1'</error>);

// A null default makes the parameter nullable.
note(null);

// null passed to a scalar parameter of PHP's own functions is only
// deprecated; an array is rejected.
strlen(null);
strlen(<error descr="Pass string to $string of strlen(), not array.">[]</error>);
