<?php
namespace Shop;

use Psr\Clock\ClockInterface as Clock;
use Psr\Link\LinkInterface as linkinterface;
use Psr\Http\Message;
use Psr\Container\{ContainerInterface as Container, NotFoundExceptionInterface};

class Cart {
    use \Shop\Traits\Totals { total as total; }
    use Countable;
}

$fn = function () use ($x) {};
