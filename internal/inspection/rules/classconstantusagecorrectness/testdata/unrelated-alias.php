<?php

namespace Shop\Model;

use Shop\Model\Basket as ModelBasket;

class Basket {}

class Order
{
    protected static string $basketClass = Basket::class;

    public function make(): ModelBasket
    {
        return new ModelBasket();
    }
}
