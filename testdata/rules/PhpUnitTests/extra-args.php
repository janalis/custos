<?php
class ExtraArgsTest
{
    public function testExtra($rows, $n, $list, $obj)
    {
        <weak_warning descr="Use 'assertEmpty()' instead.">$this->assertTrue(empty($rows), 'm', 3)</weak_warning>;
        <weak_warning descr="Use 'assertIsInt()' instead.">$this->assertTrue(is_int($n), 'm', $flag)</weak_warning>;
        <weak_warning descr="Use 'assertInstanceOf()' instead.">$this->assertTrue($obj instanceof \Countable, 'm', 3)</weak_warning>;
        <weak_warning descr="Use 'assertCount()' instead.">$this->assertSame(2, count($list), 'm', 4, 5)</weak_warning>;
    }

    public function testCase($rows, $n, $list)
    {
        <weak_warning descr="Use 'assertCount()' instead.">$this->assertSame(2, Count($list))</weak_warning>;
        <weak_warning descr="Use 'assertIsInt()' instead.">$this->assertTrue(IS_INT($n))</weak_warning>;
        <weak_warning descr="Use 'assertContains()' instead.">$this->assertTrue(\In_Array($n, $list))</weak_warning>;
    }
}
