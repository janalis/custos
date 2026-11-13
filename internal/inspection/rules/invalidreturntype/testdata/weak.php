<?php
// No strict_types: PHP coerces scalars, but never arrays, objects or null.
namespace Shop;

class Order {}

function total(): int { return '1200'; }
function ratio(): float { return 1; }
function flag(): bool { return 'yes'; }
function lines(): int { return <error descr="Return int from Shop\lines(), not array.">[]</error>; }
function missing(): int { return <error descr="Return int from Shop\missing(), not null.">null</error>; }
function order(): string { return new Order(); } // __toString() may convert it
function count_of(): Order { return <error descr="Return Order from Shop\count_of(), not string.">'order-1'</error>; }
