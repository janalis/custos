<?php
declare(strict_types=1);

namespace Shop;

// A write into an offset of false makes it an array (deprecated from
// PHP 8.1): the value is judged as that array, not as false.
final class Basket
{
    public array $items = [];
    public int $count = 0;
    private static array|false $cache = false;

    public function fill(): void
    {
        $list = false;
        $list[] = 1;
        $this->items = $list;
        $this->count = <error descr="Assign int to Shop\Basket::$count, not int[].">$list</error>;
    }

    public function lazy(): void
    {
        if (self::$cache === false) {
            self::$cache['k'] = 1;
            $this->items = self::$cache;
            $this->count = <error descr="Assign int to Shop\Basket::$count, not array.">self::$cache</error>;
        }
    }
}
