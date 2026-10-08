<?php
namespace Shop\Mail {
    class Courier {
        public function dispatch() {}
    }
}

namespace Shop\Tests\Mail {
    /**
     * @coversDefaultClass Shop\Mail\Courier
     */
    class CourierTest {
        /**
         * PHPUnit reads tag names as fully qualified.
         * @covers ::dispatch
         * @covers Shop\Mail\Courier::dispatch
         */
        public function testDispatches() {}

        /** @covers Mail\Missing::dispatch */
        public function <error descr="The @covers target 'Mail\Missing::dispatch' cannot be resolved.">testMissing</error>() {}
    }
}
