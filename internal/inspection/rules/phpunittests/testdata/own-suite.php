<?php

namespace PHPUnit\Framework {
    // A PHPUnit base: its suggestions mean what the rule expects.
    class Checks
    {
        protected function assertTrue($value) {}
        protected function assertSame($expected, $actual) {}
    }
}

namespace {
    // E12: a test base of its own declares the assertion but not the
    // suggestion, or declares the suggestion itself.
    class Suite
    {
        protected function assertTrue($value) {}
        protected static function assertFalse($value) {}
    }

    class StrictSuite
    {
        protected function assertTrue($value) {}
        protected function assertEquals($expected, $actual) {}
    }

    class OwnSuiteTest extends Suite
    {
        public function testIt($a, $b)
        {
            $this->assertTrue($a === 0);
            self::assertFalse($a !== $b);
        }
    }

    class StrictSuiteTest extends StrictSuite
    {
        public function testIt($a, $b)
        {
            $this->assertTrue($a == $b);
        }
    }

    class ChecksTest extends PHPUnit\Framework\Checks
    {
        public function testIt($a)
        {
            <weak_warning descr="Use 'assertSame()' instead.">$this->assertTrue($a === 0)</weak_warning>;
        }
    }
}
