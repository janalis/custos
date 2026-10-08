<?php
namespace App;

use Shop\Cart as Basket;

class MoreTest
{
    public function testMore($a, $b, $e, $f, $s, $list, ?string $str)
    {
        static::assertFalse($a);
        $this->assertNotTrue($a, 'm');
        $this->assertNull($a);
        $this->assertIsNotArray($a, 'm');
        $this->assertInstanceOf(Basket::class, $e);
        $this->assertNotInstanceOf($f, $e);
        $this->assertInstanceOf(\Shop\Cart::class, $e);
        $this->assertFileExists($f, 'm');
        $this->assertCount(2, $list);
        $this->assertRegExp('/a/', $s);
        $this->assertNotRegExp('/a/', $s);
        $this->assertRegExp('/a/', $s);
        $this->assertFileEquals($a, $b, 'm');
        $this->assertEmpty($str, 'm', 3);
        $this->assertNotEquals($a, $b);
        $this->assertNotSame($a, $b, 'm');

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
