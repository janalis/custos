<?php
class OffTest
{
    public function testOff($a, $list)
    {
        <weak_warning descr="Use 'assertNotTrue()' instead.">$this->assertTrue(!$a)</weak_warning>;
        $this->assertSame(2, count($list));
        $this->assertTrue(empty($a));
        $this->expects($this->exactly(1));
        $this->will($this->returnValue(1));
    }
}
