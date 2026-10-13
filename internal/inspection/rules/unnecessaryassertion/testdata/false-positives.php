<?php

namespace Shop;

class Cart {}
class Bag extends Cart {}

function made(): Cart { return new Cart(); }
/** @return Cart */
function docOnly() { return new Cart(); }

abstract class CartTest
{
    abstract protected function bag(): Bag;

    public function testTyped($param)
    {
        $this->assertInstanceOf(Cart::class, $this->bag());
        $this->assertInstanceOf('Shop\Bag', $this->bag());
        $this->assertInstanceOf(Cart::class, docOnly());
        $this->assertNotNull(made());
        $this->assertNull(made());
        $this->assertInstanceOf(Cart::class, $param);
        $x = made();
        $x = docOnly();
        $this->assertInstanceOf(Cart::class, $x);
        $this->assertNull();
        $double = $this->createMock(Cart::class);
        $double->expects(any())->method('total');
        $double->expects($matcher);
        $double->expects($this->never());
    }

    abstract protected function shelf(): ?Shelf;

    public function testNullSafe(?Shelf $shelf)
    {
        $found = $this->shelf()?->bag();
        $this->assertInstanceOf(Bag::class, $found);
        $this->assertInstanceOf(Bag::class, $shelf?->bag());
        $this->assertInstanceOf(Bag::class, $shelf?->owner->bag());
        $this->assertInstanceOf(Bag::class, $shelf?->owner->owner->bag());
        $this->assertInstanceOf(Bag::class, $shelf?->bags[0]->bag());
    }
}

final class Shelf
{
    public Shelf $owner;
    /** @var list<Shelf> */
    public array $bags = [];
    public function bag(): Bag { return new Bag(); }
}
