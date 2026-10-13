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
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertNull($this->flush())</weak_warning>;
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">self::assertEmpty($this->flush())</weak_warning>;
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertInstanceOf(Cart::class, $this->cart())</weak_warning>;
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertInternalType('object', $this->cart())</weak_warning>;
        $this->assertInstanceOf(\ArrayObject::class, $this->cart());
        $this->assertInstanceOf(Cart::class, $this->maybeCart());
        $this->assertNull($this->cart());
        $this->assertNull($this->anything());

        $made = $this->cart();
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertInstanceOf(Cart::class, $made)</weak_warning>;

        $either = $flag ? $this->cart() : $this->flush();
        $this->assertNull($either);
    }

    public function testMatcher()
    {
        $double = $this->createMock(Cart::class);
        $double->expects(<weak_warning descr="expects(any()) verifies nothing; drop the expects() call.">$this->any()</weak_warning>)->method('total')->willReturn(9);
        $double
            ->expects(<weak_warning descr="expects(any()) verifies nothing; drop the expects() call.">self::any()</weak_warning>)
            ->method('count');
        $double->expects($this->once())->method('total');
    }
}
