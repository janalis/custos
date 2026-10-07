<?php
// An unparsable PHP_UNIT_VERSION falls back to PHPUnit 8.0.
final class LedgerTest
{
    public function testLedger($m, $x)
    {
        $this->assertEquals(1.0, $x, '', <weak_warning descr="PHPUnit 8.0 deprecated the delta argument; call assertEqualsWithDelta() instead.">0.1</weak_warning>);
        $this->assertEquals(1.0, $x, message: 'differs');
        $this->assertFileNotExists('/tmp/x');
        $this->$m(1, 2, '', 0.1);
    }
}
