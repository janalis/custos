<?php
final class OldTest
{
    public function testIt()
    {
        $this->assertEquals(1, 2, '', 0.1);
        $this->assertFileNotExists('/tmp/x');
    }
}
