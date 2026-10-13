<?php
// No PHP_UNIT_VERSION configured: the indexed PHPUnit is a 7.x one (no
// assertIsInt), so nothing is deprecated yet.
namespace PHPUnit\Framework {
    abstract class Assert {
        public static function assertEquals($e, $a, $m = '', $delta = 0.0) {}
        public static function assertInternalType($t, $v, $m = '') {}
        public static function assertFileNotExists($f, $m = '') {}
    }
}

namespace {
    final class LegacyTest extends \PHPUnit\Framework\Assert
    {
        public function testIt()
        {
            $this->assertEquals(1.0, 1.05, '', 0.1);
            $this->assertFileNotExists('/tmp/x');
        }
    }
}
