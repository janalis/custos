<?php

class Meter
{
    public function read() {}
}

class MeterTest
{
    public function testRead()
    {
        $stub = $this->GetMockBuilder(Meter::class)->GETMOCK();
        $stub->Method('read')->will($this->ReturnValue(1));
        $stub->METHOD('write');
    }
}
