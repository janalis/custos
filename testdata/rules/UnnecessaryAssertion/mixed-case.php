<?php

namespace Shop;

class Basket {}

abstract class BasketTest
{
    abstract protected function clear(): void;
    abstract protected function basket(): Basket;

    public function testCase()
    {
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->AssertNull($this->clear())</weak_warning>;
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">self::assertinstanceof(Basket::class, $this->basket())</weak_warning>;
        $mock = $this->createMock(Basket::class);
        $mock->Expects(<weak_warning descr="expects(any()) verifies nothing; drop the expects() call.">$this->ANY()</weak_warning>)->method('total');
    }
}
