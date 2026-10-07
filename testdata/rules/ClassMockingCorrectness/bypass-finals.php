<?php

namespace DG {
    // dg/bypass-finals is installed: `final` is stripped when classes load.
    final class BypassFinals { public static function enable() {} }
}

namespace PHPUnit\Framework {
    class TestCase
    {
        public function getMockBuilder($c) {}
        public function createMock($c)     {}
    }
}

namespace {
    final class Clock {}
    trait Stamped {}
    final class Mailer { public function __construct($transport) {} }

    class ClockTest extends \PHPUnit\Framework\TestCase
    {
        public function testDoubles()
        {
            $this->createMock(Clock::class);
            $this->getMockBuilder('\Clock');
            $this->createMock(<error descr="Traits cannot be mocked this way.">Stamped::class</error>);
            $this->getMockBuilder(<error descr="Mocked constructor needs arguments; pass them or disable the constructor.">Mailer::class</error>)->getMock();
        }
    }
}
