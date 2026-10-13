<?php

namespace PHPUnit\Framework {
    class TestCase
    {
        public function getMockBuilder($c)  {}
        public function createMock($c)      {}
        public function getMockForTrait($c) {}
    }
}

namespace {
    final class Vault {}
    trait Audited {}
    class Ledger { public function __construct($db) {} }

    class VaultTest extends \PHPUnit\Framework\TestCase
    {
        public function testCase()
        {
            $this->CreateMock(<error descr="Final classes cannot be mocked.">Vault::class</error>);
            $this->CREATEMOCK(<error descr="Traits cannot be mocked this way.">Audited::class</error>);
            $this->getmockfortrait(<error descr="This factory expects a trait.">Ledger::class</error>);
            $this->GetMockBuilder(<error descr="Mocked constructor needs arguments; pass them or disable the constructor.">Ledger::class</error>)->GETMOCK();
            $this->createmock(Ledger::class);
        }
    }
}
