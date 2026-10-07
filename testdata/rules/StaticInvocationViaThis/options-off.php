<?php
namespace PHPUnit\Framework;

class Assert { public static function assertTrue($v) {} }
class CheckoutTest extends Assert {
    public function testTotal() { <warning descr="Static method assertTrue() called through $this; use self::assertTrue().">$this</warning>->assertTrue(true); }
}
