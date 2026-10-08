<?php
class ExtraArgsTest
{
    public function testExtra(array $rows, $n, $list, $obj)
    {
        $this->assertEmpty($rows, 'm', 3);
        $this->assertIsInt($n, 'm', $flag);
        $this->assertInstanceOf(\Countable::class, $obj, 'm', 3);
        $this->assertCount(2, $list, 'm', 4, 5);
    }

    public function testCase($rows, $n, $list)
    {
        $this->assertCount(2, $list);
        $this->assertIsInt($n);
        $this->assertContains($n, $list);
    }
}
