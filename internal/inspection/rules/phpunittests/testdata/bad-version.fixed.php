<?php
// An unparsable PHP_UNIT_VERSION is no version: PHPUnit 8 is assumed, and
// names that changed in 9.1 are not fixed.
class RegexTest
{
    public function testRegex($token, $v)
    {
        $this->assertSame(1, preg_match('/^[a-f]+$/', $token));
        $this->assertIsInt($v);
    }
}
