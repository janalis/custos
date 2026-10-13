<?php
// No PHP_UNIT_VERSION configured: the indexed Assert class decides.
namespace PHPUnit\Framework {
    abstract class Assert {
        public static function assertIsInt($v, $m = '') {}
        public static function assertInternalType($t, $v, $m = "") {}
    }
    abstract class TestCase extends Assert {}
}

namespace {
    class ValueTest extends \PHPUnit\Framework\TestCase
    {
        public function testValue($v, $list, $m)
        {
            $this->assertIsInt($v);
            $this->assertContains($v, $list);
            $this->$m($v);
        }
    }
}
