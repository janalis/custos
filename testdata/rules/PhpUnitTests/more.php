<?php
namespace App;

use Shop\Cart as Basket;

class MoreTest
{
    public function testMore($a, $b, $e, $f, $s, $list, ?string $str)
    {
        <weak_warning descr="Use 'assertFalse()' instead.">static::assertSame(false, $a)</weak_warning>;
        <weak_warning descr="Use 'assertNotTrue()' instead.">$this->assertNotSame($a, TRUE, 'm')</weak_warning>;
        <weak_warning descr="Use 'assertNull()' instead.">$this->assertTrue(is_null($a))</weak_warning>;
        <weak_warning descr="Use 'assertIsNotArray()' instead.">$this->assertFalse(is_array($a), 'm')</weak_warning>;
        <weak_warning descr="Use 'assertInstanceOf()' instead.">$this->assertTrue($e instanceof Basket)</weak_warning>;
        <weak_warning descr="Use 'assertNotInstanceOf()' instead.">$this->assertFalse($e instanceof $f)</weak_warning>;
        <weak_warning descr="Use 'assertInstanceOf()' instead.">$this->assertSame(get_class($e), 'Shop\\Cart')</weak_warning>;
        <weak_warning descr="Use 'assertFileExists()' instead.">$this->assertTrue(file_exists($f), 'm')</weak_warning>;
        <weak_warning descr="Use 'assertCount()' instead.">$this->assertSame(2, count($list))</weak_warning>;
        <weak_warning descr="Use 'assertRegExp()' instead.">$this->assertSame(1, preg_match('/a/', $s))</weak_warning>;
        <weak_warning descr="Use 'assertNotRegExp()' instead.">$this->assertEquals(0, preg_match('/a/', $s))</weak_warning>;
        <weak_warning descr="Use 'assertRegExp()' instead.">$this->assertTrue(preg_match('/a/', $s) > 0)</weak_warning>;
        <weak_warning descr="Use 'assertFileEquals()' instead.">$this->assertStringEqualsFile($a, file_get_contents($b), 'm')</weak_warning>;
        <weak_warning descr="Use 'assertEmpty()' instead.">$this->assertTrue(empty($str), 'm', 3)</weak_warning>;
        <weak_warning descr="Use 'assertNotEquals()' instead.">$this->assertFalse(($a == $b))</weak_warning>;
        <weak_warning descr="Use 'assertNotSame()' instead.">$this->assertNotTrue($a === $b, 'm')</weak_warning>;

        $this->assertTrue(-$a);
        $this->assertTrue(is_integer($a));
        $this->assertTrue((empty($a)));
        $this->assertSame(count($list, COUNT_RECURSIVE), 2);
        $this->assertSame(2, count($list, COUNT_RECURSIVE));
        $this->assertSame(1, preg_match('/a/', $s, $m));
        $this->assertTrue(preg_match('/a/', $s) >= 0);
        $this->assertNotTrue(preg_match('/a/', $s) > 0);
        $this->assertSame('abc', get_class($e));
        $this->assertTrue(Count($list) == 0 ? true : false);
        $this->assertEquals(1, 2);
        $mock->expects($this->exactly(01))->method('x');
        $mock->will($this->returnValue(1, 2));
        $mock->will(returnValue(1));
    }

    public function assertFileEquals($a, $b)
    {
        $this->assertSame(file_get_contents($a), file_get_contents($b));
    }
}
