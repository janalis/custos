<?php
namespace Shop;

use Other\Basket as OtherBasket;

class Basket
{
    public $prop = Basket::class;

    public function check($x)
    {
        $a = basket::class;
        $b = \shop\Basket::class;
        $c = new OtherBasket();
        $d = \Other\Basket::class;
        $e = static::class;
        $f = parent::x();
        $g = self::class;
        $h = __CLASS__;
        /** @var Basket $i */
        $i = make();
        $j = Basket();
        $k = Basket;
    }
}

namespace Other;

class Shelf
{
    public function get(): \Shop\Basket { return new Basket(); }
}
