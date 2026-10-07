<?php

namespace Shop;

class Basket {}

abstract class BasketTest
{
    abstract protected function clear(): void;
    abstract protected function basket(): Basket;

    public function testCase()
    {
        $this->AssertNull($this->clear());
        self::assertinstanceof(Basket::class, $this->basket());
        $mock = $this->createMock(Basket::class);
        $mock->method('total');
    }
}
