<?php

namespace PhpSpec {
    class ObjectBehavior {}
}

namespace PHPUnit\Framework {
    class TestCase
    {
        public function getMockBuilder($c)          {}
        public function getMockForTrait($c)         {}
        public function getMockForAbstractClass($c) {}
        public function getMockClass($c)            {}
        public function createMock($c)              {}
    }
}

namespace {
    class PHPUnit_Framework_MockObject_Generator
    {
        public function getMock($c) {}
    }

    final class Sealed { public function __construct($a) {} const LIMIT = 1; }
    abstract class Blueprint {}
    trait Mixin {}
    interface Contract {}
    class Gateway { public function __construct($dsn, $retries = 3) {} }
    class Relaxed { public function __construct($dsn = '') {} }
    class Spread { public function __construct(...$parts) {} }
    class Unrelated { public function createMock($c) {} }

    class OrderTest extends \PHPUnit\Framework\TestCase
    {
        public function testDoubles(Unrelated $other, $name)
        {
            $this->createMock(<error descr="Final classes cannot be mocked.">Sealed::class</error>);
            $this->createMock(<error descr="Traits cannot be mocked this way.">Mixin::class</error>);
            $this->createMock(Contract::class);
            $this->createMock($name);
            $this->createMock();
            $other->createMock(Sealed::class);
            $this->getMockBuilder(<error descr="Final classes cannot be mocked.">'Sealed'</error>);
            $this->getMockBuilder(<error descr="Final classes cannot be mocked.">"\\Sealed"</error>);
            $this->getMockBuilder('Blueprint');
            $this->getMockBuilder("{$name}");
            $this->getMockBuilder(<error descr="Abstract class: build the double with getMockForAbstractClass().">Blueprint::class</error>);
            $this->getMockBuilder(Blueprint::class)->getMockForAbstractClass();
            $this->getMockBuilder(<error descr="Trait: build the double with getMockForTrait().">Mixin::class</error>);
            $this->getMockBuilder(Mixin::class)->getMockForTrait();
            $this->getMockBuilder(Contract::class);
            $this->getMockForTrait(<error descr="This factory expects a trait.">Blueprint::class</error>);
            $this->getMockForTrait(Mixin::class);
            $this->getMockForAbstractClass(<error descr="This factory expects an abstract class.">Gateway::class</error>);
            $this->getMockForAbstractClass(Contract::class);
            $this->getMockClass(<error descr="Final classes cannot be mocked.">Sealed::class</error>);
            $this->getMockBuilder(<error descr="Mocked constructor needs arguments; pass them or disable the constructor.">Gateway::class</error>)->getMock();
            $this->getMockBuilder(<error descr="Final classes cannot be mocked."><error descr="Mocked constructor needs arguments; pass them or disable the constructor.">Sealed::class</error></error>)->getMock();
            $this->getMockBuilder(Relaxed::class)->getMock();
            $this->getMockBuilder(Spread::class)->getMock();
            $this->getMockBuilder(Gateway::class)->disableOriginalConstructor()->getMock();
            parent::createMock(<error descr="Final classes cannot be mocked.">Sealed::class</error>);
            $m = 'createMock';
            $this->$m(Sealed::class);
            $this->createMock(...[Sealed::class]);
            $this->createMock(Sealed::LIMIT);
            $this->createMock($other::class);
            $this->getMockBuilder(<error descr="Final classes cannot be mocked.">Sealed::class</error>)->$m();
        }
    }

    $generator = new PHPUnit_Framework_MockObject_Generator();
    $generator->getMock(<error descr="Final classes cannot be mocked.">Sealed::class</error>);
    $generator->getMock(ArrayObject::class);

    class GatewaySpec extends \PhpSpec\ObjectBehavior
    {
        function it_charges(<error descr="Final classes cannot be mocked.">Sealed</error> $card, Relaxed $log, ?<error descr="Final classes cannot be mocked.">Sealed</error> $maybe, int $n, $limit = Sealed::LIMIT) {}
        private $state;
        function it_unions(<error descr="Final classes cannot be mocked.">Sealed</error>|Relaxed $first) {}
        function it_limits($max = Sealed::LIMIT, Relaxed|<error descr="Final classes cannot be mocked.">Sealed</error> $either = Sealed::DEFAULT) {}
    }

    class IndirectSpec extends GatewaySpec
    {
        function it_skips(Sealed $card) {}
    }
}
