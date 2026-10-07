<?php
namespace PHPUnit\Framework;

class Assert { public static function assertTrue($v) {} }
class CheckoutTest extends Assert {
    public function testTotal() { <warning descr="Static method assertTrue() called through $this; use static::assertTrue().">$this</warning>->assertTrue(true); }
}
