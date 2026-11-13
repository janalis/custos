<?php
// An unparsable PHP_UNIT_VERSION is no version: PHPUnit 8 is assumed, and
// names that changed in 9.1 are not fixed.
class RegexTest
{
    public function testRegex($token, $v)
    {
        <weak_warning descr="Use 'assertMatchesRegularExpression()' ('assertRegExp()' before PHPUnit 9.1) instead.">$this->assertSame(1, preg_match('/^[a-f]+$/', $token))</weak_warning>;
        <weak_warning descr="Use 'assertIsInt()' instead.">$this->assertTrue(is_int($v))</weak_warning>;
    }
}
