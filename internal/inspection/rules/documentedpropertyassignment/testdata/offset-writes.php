<?php
declare(strict_types=1);

namespace Shop;

// A write into an offset of false makes it an array (deprecated from
// PHP 8.1): the value is judged as that array, not as false.
final class Basket
{
    public array $items = [];
    /** @var int[] */
    public $ids = [];
    /** @var int */
    public $count = 0;
    /** @var array|false */
    private $cache = false;
    /** @var array */
    public $copy = [];

    public function fill(): void
    {
        $list = false;
        $list[] = 1;
        $this->items = $list;
        $this->ids = $list;
    }

    public function fillStatic(): void
    {
        $rows = false;
        $rows['k'] = 1;
        $this->items = $rows;
    }

    // The array is still rejected where it does not fit.
    public function tally(): void
    {
        $list = false;
        $list[] = 1;
        $this->count = <warning descr="Assign int to Shop\Basket::$count, not int[].">$list</warning>;
    }

    // So is a property only PHPDoc types, once a test shows it held false.
    public function lazy(): void
    {
        if ($this->cache === false) {
            $this->cache[] = 1;
            $this->copy = $this->cache;
            $this->count = <warning descr="Assign int to Shop\Basket::$count, not array.">$this->cache</warning>;
        }
    }
}
