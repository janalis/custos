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
        $stub->Method('read')-><warning descr="The stub object is returned as-is here; use '->will(...)'.">WillReturn</warning>($this->ReturnValue(1));
        $stub->METHOD(<error descr="The mocked class has no such method.">'write'</error>);
    }
}
