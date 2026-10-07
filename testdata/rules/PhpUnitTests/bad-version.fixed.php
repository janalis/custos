<?php
// An unparsable PHP_UNIT_VERSION falls back to PHPUnit 8.
class RegexTest
{
    public function testRegex($token, $v)
    {
        $this->assertRegExp('/^[a-f]+$/', $token);
        $this->assertIsInt($v);
    }
}
