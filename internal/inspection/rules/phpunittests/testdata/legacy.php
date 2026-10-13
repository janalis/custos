<?php
class LegacyTest
{
    public function testTypes(int $qty, $raw, array $list)
    {
        <weak_warning descr="Use 'assertInternalType('int', …)' instead.">$this->assertTrue(is_int($qty))</weak_warning>;
        <weak_warning descr="Use 'assertNotInternalType('iterable', …)' instead.">$this->assertFalse(is_iterable($raw), 'iter')</weak_warning>;
        <weak_warning descr="Loose comparison; use 'assertSame()' instead.">$this->assertEquals(12, $qty)</weak_warning>;
        <weak_warning descr="Loose comparison; use 'assertNotSame()' instead.">$this->assertNotEquals('7', 3.5, 'odd')</weak_warning>;
        <weak_warning descr="Use 'assertNotContains()' instead.">$this->assertNotTrue(in_array($qty, $list))</weak_warning>;
        $this->assertEquals($list, []);
        $this->assertEquals($raw, 12);
    }
}
