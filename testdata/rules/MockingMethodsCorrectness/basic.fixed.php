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

        $stub->method('balance')->will($this->returnValue(42));
        $stub->method('balance')->will(self::returnCallback('strtoupper'));
        $stub->method('balance')->will($this->returnValue(42));
        $stub->method('balance')->willReturn(42);

        $stub->method('seal')->willReturn(true);
        $stub->method('refund')->willReturn(0);
        $stub->expects($this->once())->method("refund");
        $stub->method('virtualTotal')->willReturn(7);
        $stub->method('BALANCE')->willReturn(1);
        $stub->method($dynamic)->willReturn(1);

        $this->getMockBuilder(Ledger::class)->getMock()->method('seal');
        $this->createMock(Ledger::class)->method('refund');
    }
}
