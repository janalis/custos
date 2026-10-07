<?php

class Registry
{
    public function getManager() {}
}

class RegistryTest extends \PHPUnit\Framework\TestCase
{
    public function testAdded()
    {
        $doctrine = $this->getMockBuilder(\stdClass::class)
            ->addMethods(['getManager', 'getRepository'])
            ->getMock();
        $doctrine->method('getRepository')->willReturn(null);
        $doctrine->method('GetManager')->willReturn(null);

        $registry = $this->getMockBuilder(Registry::class)
            ->addMethods(['findOneByNom'])
            ->getMock();
        $registry->method('findOneByNom')->willReturn(null);
        $registry->method(<error descr="The mocked class has no such method.">'findAll'</error>)->willReturn(null);
    }
}
