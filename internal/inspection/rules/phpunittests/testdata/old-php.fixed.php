<?php
namespace App;

final class Cart {}

class OldTest
{
    public function testOld($e)
    {
        $this->assertInstanceOf('\\App\\Cart', $e);
        $this->assertNotInstanceOf('\\App\\Cart', $e);
    }
}
