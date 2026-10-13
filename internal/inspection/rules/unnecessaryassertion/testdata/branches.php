<?php

namespace Shop;

class Cart
{
    public static function make(): static { return new static(); }
    public function me(): static { return $this; }
    public function flush(): void {}
}

abstract class BaseTest
{
    abstract protected function base(): BaseTest;
}

abstract class BranchesTest extends BaseTest
{
    abstract protected function self(): BranchesTest;
    abstract protected function flag(): bool;
    abstract protected function nothing(): void;
    abstract protected function res(): resource;
    abstract protected function cb(): callable;
    abstract protected function closure(): \Closure;
    abstract protected function any(): mixed;
    /** @return int[] */
    abstract protected function ids(): array;

    public function testBranches(Cart $cart, $unknown, string $m)
    {
        $this->$m();
        $this->assertNull(nope());
        $this->assertNull($this->$m());
        $this->assertNull($unknown->flush());
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertNull($cart->flush())</weak_warning>;
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertInstanceOf(Cart::class, Cart::make())</weak_warning>;
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertInstanceOf(Cart::class, $cart::make())</weak_warning>;
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertInstanceOf(Cart::class, $cart->me())</weak_warning>;
        $this->assertNull($unknown::make());
        $this->assertNull(Cart::missing());
        $this->assertNull(null);
        $this->assertInstanceOf(Missing::class, $cart->me());
        $this->assertInstanceOf(Cart::NAME, $cart->me());
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertInstanceOf(self::class, $this->self())</weak_warning>;
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertInstanceOf(parent::class, $this->base())</weak_warning>;
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertInternalType('bool', $this->flag())</weak_warning>;
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertInternalType('null', $this->nothing())</weak_warning>;
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertInternalType('callable', $this->cb())</weak_warning>;
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertInternalType('callable', $this->closure())</weak_warning>;
        $this->assertInternalType('resource', $this->res());
        $this->assertInternalType('int', $this->any());
        $this->assertInternalType('array', $this->ids());
        $this->assertInternalType('string', 'literal');
        $double = $this->createMock(Cart::class);
        $double->expects();
        $double->expects($this->any(), 2);
    }
}

abstract class NoParentTest
{
    public function testParent()
    {
        $this->assertNull(parent::flush());
    }
}
