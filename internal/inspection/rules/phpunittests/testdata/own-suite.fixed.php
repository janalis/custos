<?php

// E12: a test base of its own declares the assertion but not the suggestion.
class Suite
{
    protected function assertTrue($value) {}
    protected static function assertFalse($value) {}
}

class Checks
{
    protected function assertTrue($value) {}
    protected function assertSame($expected, $actual) {}
}

class OwnSuiteTest extends Suite
{
    public function testIt($a, $b)
    {
        $this->assertTrue($a === 0);
        self::assertFalse($a !== $b);
    }
}

class ChecksTest extends Checks
{
    public function testIt($a)
    {
        $this->assertSame($a, 0);
    }
}
