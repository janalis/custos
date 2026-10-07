<?php
// An unparsable PHP_UNIT_VERSION falls back to PHPUnit 8.
class RegexTest
{
    public function testRegex($token, $v)
    {
        <weak_warning descr="Use 'assertRegExp()' instead.">$this->assertSame(1, preg_match('/^[a-f]+$/', $token))</weak_warning>;
        <weak_warning descr="Use 'assertIsInt()' instead.">$this->assertTrue(is_int($v))</weak_warning>;
    }
}
