<?php
namespace PHPUnit\Framework {
    class Assert { public static function assertTrue($v) {} }
    class CheckoutTest extends Assert {
        public function testTotal() { $this->assertTrue(true); }
    }
}
namespace App {
    class Unknown {
        public function run($x) { return $x->make() . $this->missing(); }
    }
}
