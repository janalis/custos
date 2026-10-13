<?php

class Ledger
{
    public function balance() {}
    final public function seal() {}
}

class LedgerTest
{
    public function testDoubles()
    {
        $stub = $this->getMockBuilder(Ledger::class)
            ->setMethods(['virtualTotal'])
            ->getMock();

        $stub->method('balance')-><warning descr="The stub object is returned as-is here; use '->will(...)'.">willReturn</warning>($this->returnValue(42));
        $stub->method('balance')-><warning descr="The stub object is returned as-is here; use '->will(...)'.">willReturn</warning>(self::returnCallback('strtoupper'));
        $stub->method('balance')->will($this->returnValue(42));
        $stub->method('balance')->willReturn(42);

        $stub->method(<error descr="Final methods cannot be mocked.">'seal'</error>)->willReturn(true);
        $stub->method(<error descr="The mocked class has no such method.">'refund'</error>)->willReturn(0);
        $stub->expects($this->once())->method(<error descr="The mocked class has no such method.">"refund"</error>);
        $stub->method('virtualTotal')->willReturn(7);
        $stub->method('BALANCE')->willReturn(1);
        $stub->method($dynamic)->willReturn(1);

        $this->getMockBuilder(Ledger::class)->getMock()->method(<error descr="Final methods cannot be mocked.">'seal'</error>);
        $this->createMock(Ledger::class)->method('refund');
    }
}
