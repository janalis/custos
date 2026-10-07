<?php
class Widget {}

function assertFileEquals($a, $b)
{
    $this->assertSame(file_get_contents($a), file_get_contents($b));
}

function helperCompare($a, $b)
{
    <weak_warning descr="Use 'assertFileEquals()' instead.">$this->assertSame(file_get_contents($a), file_get_contents($b))</weak_warning>;
}

class WidgetTest
{
    public function testBranches($x, $s, array $list, $stub)
    {
        $this->assertTrue(...$list);
        <weak_warning descr="Use 'assertNotFalse()' instead.">$this->assertFalse(!$x)</weak_warning>;
        <weak_warning descr="Use 'assertNotTrue()' instead.">$this->assertTrue(!$x, 'message')</weak_warning>;
        <weak_warning descr="Use 'assertNull()' instead.">$this->assertSame(PHP_EOL, null)</weak_warning>;
        $this->assertTrue(is_int());
        $this->assertSame(get_class($x));
        $this->assertSame('Widget\Thing', get_class());
        $this->assertSame(count($list));
        $this->assertTrue();
        $this->assertTrue(preg_match('/x/', $s, $m) > 0);
        $check = function () use ($a, $b) {
            <weak_warning descr="Use 'assertFileEquals()' instead.">$this->assertSame(file_get_contents($a), file_get_contents($b))</weak_warning>;
        };
        $stub->expects();
        $stub->expects($this->once());
        $stub->will($value);
        $stub->will();
    }
}
