<?php
// No PHP_UNIT_VERSION configured: the version follows the indexed PHPUnit,
// whose Assert class has the 9.1 regex assertions.
namespace PHPUnit\Framework {
    abstract class Assert {
        public static function assertMatchesRegularExpression($p, $s, $m = '') {}
        public static function assertDoesNotMatchRegularExpression($p, $s, $m = '') {}
        public static function assertIsInt($v, $m = '') {}
        public static function assertSame($e, $a, $m = '') {}
    }
    abstract class TestCase extends Assert {}
}

namespace {
    class TokenTest extends \PHPUnit\Framework\TestCase
    {
        public function testToken($token)
        {
            <weak_warning descr="Use 'assertMatchesRegularExpression()' instead.">$this->assertSame(1, preg_match('/^[a-f0-9]+$/', $token))</weak_warning>;
            <weak_warning descr="Use 'assertDoesNotMatchRegularExpression()' instead.">$this->assertNotSame(1, preg_match('/\s/', $token))</weak_warning>;
        }
    }
}
