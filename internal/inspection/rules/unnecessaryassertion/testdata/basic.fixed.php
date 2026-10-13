<?php

namespace Shop;

class Cart {}

abstract class CartTest
{
    abstract protected function flush(): void;
    abstract protected function cart(): Cart;
    abstract protected function maybeCart(): ?Cart;
    abstract protected function anything();

    public function testTyped(bool $flag)
    {
        $this->assertNull($this->flush());
        self::assertEmpty($this->flush());
        $this->assertInstanceOf(Cart::class, $this->cart());
        $this->assertInternalType('object', $this->cart());
        $this->assertInstanceOf(\ArrayObject::class, $this->cart());
        $this->assertInstanceOf(Cart::class, $this->maybeCart());
        $this->assertNull($this->cart());
        $this->assertNull($this->anything());

        $made = $this->cart();
        $this->assertInstanceOf(Cart::class, $made);

        $either = $flag ? $this->cart() : $this->flush();
        $this->assertNull($either);
    }

    public function testMatcher()
    {
        $double = $this->createMock(Cart::class);
        $double->method('total')->willReturn(9);
        $double
            ->method('count');
        $double->expects($this->once())->method('total');
    }
}
