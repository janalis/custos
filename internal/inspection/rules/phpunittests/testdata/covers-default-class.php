<?php
namespace Shop\Mail {
    class Courier {
        public function __construct($transport) {}
        public function dispatch() {}
    }
}

namespace Shop\Tests {
    /**
     * @coversDefaultClass \Shop\Mail\Courier
     */
    class CourierTest {
        /**
         * @covers ::__construct
         * @covers ::dispatch
         */
        public function testDispatches() {}

        /** @covers ::strlen */
        public function testFallsBackToFunctions() {}

        /** @covers ::recall */
        public function <error descr="The @covers target '::recall' cannot be resolved.">testRecalls</error>() {}
    }

    class PlainTest {
        /** @covers ::dispatch */
        public function <error descr="The @covers target '::dispatch' cannot be resolved.">testNoDefaultClass</error>() {}
    }
}
