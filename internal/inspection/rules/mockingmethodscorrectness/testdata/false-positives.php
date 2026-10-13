<?php

class Ledger
{
    public function balance() {}
    final public function seal() {}
}

class LedgerHelper
{
    public function notATest()
    {
        $stub = $this->getMockBuilder(Ledger::class)->getMock();
        $stub->method('refund')->willReturn($this->returnValue(1));
    }
}

class LedgerTest
{
    public function testTwice($flag)
    {
        $stub = $this->getMockBuilder(Ledger::class)->getMock();
        $stub = $this->getMockBuilder(Ledger::class)->getMock();
        $stub->method('refund');
        $other = $this->getMockBuilder('Ledger')->getMock();
        $other->method('refund');
        $third = $this->getMockBuilder(Missing::class)->getMock();
        $third->method('refund');
        $this->getMockBuilder(Ledger::class)->getMock()->method('balance');
        $this->getMockBuilder(Ledger::class)->getMock()->method(self::NAME);
        $stub->method('a')->willReturn(($this->returnValue(1)));
        $stub->method('a')->willReturn(returnValue(1));
        $stub->method('a')->willReturn($this->returnSelf());
    }
}
