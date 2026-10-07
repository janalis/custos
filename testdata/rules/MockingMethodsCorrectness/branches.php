<?php

class Vault
{
    public function open() {}
    final public function seal() {}
}

class VaultTest
{
    public function testBranches(string $m, array $list)
    {
        $stub = $this->getMockBuilder(Vault::class)->getMock();
        $stub->$m('open');
        $stub->method();
        $stub->method('open', 'extra');
        $stub->method(<<<'NAME'
            open
            NAME);
        $stub->method(<error descr="Final methods cannot be mocked."><<<NAME
        seal
        NAME</error>);
        $stub->method(<error descr="The mocked class has no such method."><<<'NAME'
            close
            NAME</error>);
        $stub->method(<error descr="The mocked class has no such method."><<<'NAME'
            NAME</error>);

        $plain = $this->getMock();
        $plain->method('anything');

        $spread = $this->getMockBuilder(Vault::class)->setMethods(...$list)->getMock();
        $spread->method(<error descr="The mocked class has no such method.">'close'</error>);
        $nonArray = $this->getMockBuilder(Vault::class)->setMethods(null)->getMock();
        $nonArray->method(<error descr="The mocked class has no such method.">'close'</error>);
        $noArgs = $this->getMockBuilder()->getMock();
        $noArgs->method('close');
    }
}
