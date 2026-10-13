<?php

final class InvoiceTest
{
    public function testTotals()
    {
        $this->assertEquals(10.5, $sum, 'sum differs', <weak_warning descr="PHPUnit 8.0 deprecated the delta argument; call assertEqualsWithDelta() instead.">0.01</weak_warning>);
        self::assertNotEquals(
            [1, 2], $ids, '',
            <weak_warning descr="PHPUnit 8.0 deprecated the delta argument; call assertNotEqualsWithDelta() instead.">0</weak_warning>,
            <weak_warning descr="PHPUnit 8.0 deprecated the maxDepth argument; drop it.">$depth</weak_warning>,
            <weak_warning descr="PHPUnit 8.0 deprecated the canonicalize argument; call assertNotEqualsCanonicalizing() instead.">true</weak_warning>,
            <weak_warning descr="PHPUnit 8.0 deprecated the ignoreCase argument; call assertNotEqualsIgnoringCase() instead.">$fold = true</weak_warning>
        );
        $this->assertEquals(1.0, $x, <weak_warning descr="PHPUnit 8.0 deprecated the delta argument; call assertEqualsWithDelta() instead.">delta: 0.1</weak_warning>);
        $this->assertEquals('a', $label, 'label differs');
        $this->assertFileNotExists('/tmp/out.csv');
        $this->AssertEquals(1, 2, '', <weak_warning descr="PHPUnit 8.0 deprecated the delta argument; call assertEqualsWithDelta() instead.">0.1</weak_warning>);
        assertEquals(1, 2, '', 0.1);
    }
}
