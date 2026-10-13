<?php
namespace {
    function clean_param($v) { return $v; }
}

namespace Tests {
    class CleanTest {
        /** @covers \clean_param */
        public function testClean() {}

        /** @covers clean_param */
        public function testBare() {}

        /** @covers \no_such_function */
        public function <error descr="The @covers target '\no_such_function' cannot be resolved.">testMissing</error>() {}
    }
}
