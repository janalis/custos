<?php

namespace Shop;

class Cart {}

abstract class InternalTypeTest
{
    abstract protected function cart(): Cart;
    abstract protected function count(): int;
    abstract protected function items(): array;
    abstract protected function ratio(): float;

    public function testTypes($name)
    {
        $this->assertInternalType('array', $this->cart());
        $this->assertInternalType('string', $this->count());
        $this->assertInternalType('no-such-type', $this->count());
        $this->assertInternalType($name, $this->count());
        $this->assertInternalType('int', $this->items());
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertInternalType('object', $this->cart())</weak_warning>;
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertInternalType('integer', $this->count())</weak_warning>;
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertInternalType('numeric', $this->count())</weak_warning>;
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertInternalType('array', $this->items())</weak_warning>;
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertInternalType('iterable', $this->items())</weak_warning>;
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertInternalType("double", $this->ratio())</weak_warning>;
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertInternalType('scalar', $this->ratio())</weak_warning>;
    }
}
