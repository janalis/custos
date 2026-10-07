<?php
class OrphanTest
{
    public function testOrphan($x)
    {
        $this->assertInstanceOf('\\parent', $x);
    }
}
