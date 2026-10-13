<?php
namespace App;

final class Cart {}

class OldTest
{
    public function testOld($e)
    {
        <weak_warning descr="Use 'assertInstanceOf()' instead.">$this->assertTrue($e instanceof Cart)</weak_warning>;
        <weak_warning descr="Use 'assertNotInstanceOf()' instead.">$this->assertNotEquals('App\\Cart', get_class($e))</weak_warning>;
    }
}
