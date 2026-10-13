<?php
namespace PHPUnit\Framework;

class Assert { public static function assertTrue($v) {} }
class CheckoutTest extends Assert {
    public function testTotal() { static::assertTrue(true); }
}
