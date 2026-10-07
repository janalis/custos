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
}
