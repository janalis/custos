<?php

class Vault
{
    final public function seal() {}
}

class VaultTest
{
    public function testUnstable()
    {
        $stub = $this->getMockBuilder(Vault::class)->getMock();
        if (rand()) {
            $stub ??= $this->createMock(Vault::class);
        }
        $stub->method('seal')->willReturn(true);
        $stub->method('missing')->willReturn(0);
    }
}
