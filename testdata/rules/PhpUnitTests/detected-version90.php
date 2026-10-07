<?php
// No PHP_UNIT_VERSION configured: the indexed Assert class decides.
namespace PHPUnit\Framework {
    abstract class Assert {
        public static function assertIsInt($v, $m = '') {}

    }
    abstract class TestCase extends Assert {}
}

namespace {
    class ValueTest extends \PHPUnit\Framework\TestCase
    {
        public function testValue($v, $list, $m)
        {
            <weak_warning descr="Use 'assertIsInt()' instead.">$this->assertTrue(is_int($v))</weak_warning>;
            $this->assertTrue(in_array($v, $list));
            $this->$m($v);
        }
    }
}
