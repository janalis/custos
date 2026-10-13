<?php
class LegacyTest
{
    public function testTypes(int $qty, $raw, array $list)
    {
        $this->assertInternalType('int', $qty);
        $this->assertNotInternalType('iterable', $raw, 'iter');
        $this->assertSame(12, $qty);
        $this->assertNotSame('7', 3.5, 'odd');
        $this->assertNotContains($qty, $list);
        $this->assertEquals($list, []);
        $this->assertEquals($raw, 12);
    }
}
