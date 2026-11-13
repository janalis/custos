<?php
declare(strict_types=1);

namespace Shop;

class Address
{
    // Destructuring replaces the null the variable held before.
    public function split(string $host): void
    {
        $port = null;
        if (strpos($host, ':') !== false) {
            [$host, $port] = explode(':', $host);
            $this->port($port);
        }
        $this->port(<warning descr="Check this value for null before passing it to $port of Shop\Address::port(): it is null|string.">$port</warning>);
        $tail = null;
        list(, $tail) = explode('/', $host, 2);
        $this->port($tail);
    }

    public function port(string $port): void {}
}
