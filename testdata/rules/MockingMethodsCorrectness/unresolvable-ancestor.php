<?php

namespace Shop\Tests;

use Acme\Unknown\RemoteGateway;

class Gateway extends RemoteGateway
{
    public function ping(): bool { return true; }
}

class Sealed
{
    public function own(): int { return 1; }
}

class GatewayTest extends \PHPUnit\Framework\TestCase
{
    public function testInherited(): void
    {
        $gw = $this->getMockBuilder(Gateway::class)->getMock();
        $gw->method('send')->willReturn(true);
        $gw->method('ping')->willReturn(false);

        $sealed = $this->getMockBuilder(Sealed::class)->getMock();
        $sealed->method(<error descr="The mocked class has no such method.">'send'</error>)->willReturn(true);
    }
}
