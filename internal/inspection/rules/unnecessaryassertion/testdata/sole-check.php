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

    // A check in a branch, loop or operand without the assertion may not run.
    public function testConditionalExpectation($error, array $rows)
    {
        if ($error) {
            $this->expectException(\Exception::class);
        } else {
            $error = null;
        }
        foreach ($rows as $row) {
            $this->assertNotNull($row);
        }
        $error ?? $this->expectNotToPerformAssertions();
        $item = made();
        if ($error) {
            return;
        }
        $this->assertInstanceOf(Cart::class, $item);
    }

    // Redundant assertions are no companions: removing all would leave none.
    public function testOnlyRedundant()
    {
        $this->assertInstanceOf(Cart::class, made());
        $this->assertInstanceOf(Cart::class, made());
    }

    public function testRedundantBesideCheck($ok)
    {
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertInstanceOf(Cart::class, made())</weak_warning>;
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertInstanceOf(Cart::class, made())</weak_warning>;
        $this->assertTrue($ok);
    }

    public function testSameBranch($ok)
    {
        if ($ok) {
            $this->assertTrue($ok);
            <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertInstanceOf(Cart::class, made())</weak_warning>;
        }
    }
}

<weak_warning descr="The declared return type already guarantees this; the assertion can go.">CartTest::assertInstanceOf(Cart::class, made())</weak_warning>;
