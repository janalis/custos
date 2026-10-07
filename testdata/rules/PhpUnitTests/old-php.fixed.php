<?php
namespace App;

class OldTest
{
    public function testOld($e)
    {
        $this->assertInstanceOf('\\App\\Cart', $e);
        $this->assertNotInstanceOf('\\Shop\\Cart', $e);
    }
}
