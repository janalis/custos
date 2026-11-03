<?php

namespace Shop;

class Cart {}

function made(): Cart { return new Cart(); }

abstract class CartTest
{
    public function testOnlyCheck()
    {
        $this->assertInstanceOf(Cart::class, made());
    }

    public function testOnlyCheckInCallback()
    {
        $run = function () {
            $this->assertInstanceOf(Cart::class, made());
        };
        $run();
    }

    public function testWithExpectation()
    {
        $this->expectNotToPerformAssertions();
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertInstanceOf(Cart::class, made())</weak_warning>;
    }

    public function testWithMock()
    {
        $double = $this->createMock(Cart::class);
        $double->expects($this->once())->method('total');
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertInstanceOf(Cart::class, made())</weak_warning>;
    }
}

<weak_warning descr="The declared return type already guarantees this; the assertion can go.">CartTest::assertInstanceOf(Cart::class, made())</weak_warning>;
