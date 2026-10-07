<?php
namespace Shop\Mail {
    class Courier {
        public function dispatch() {}
    }
    function deliver() {}
}

namespace Shop\Tests {
    use Attribute;

    /** A class docblock without a default class. */
    abstract class BranchTest {
        #[Attribute] /** @covers ::nowhere */
        public function testAttributed() {}

        /**
         * @param int $x mail me @ home
         * @covers
         * @covers 1Foo::bar
         * @covers \Shop\Mail\
         * @covers \Shop\1Mail
         * @covers \Shop\Mail\Courier::dis-patch
         * @covers \::dispatch
         * @dataProvider 123
         */
        public function testIgnoredValues($x) {}

        /** @covers \Shop\Mail\Courier:: */
        public function testClassWithColons() {}

        /** @covers \Shop\Mail\Courier */
        public function testWholeClass() {}

        /** @covers \Shop\Mail\deliver */
        public function testFunctionAsClass() {}

        /** @covers \Shop\Mail\Missing */
        public function testMissingClass() {}

        /** @dataProvider ::provider */
        public function testEmptyClassPart($v) {}

        /** @depends \Shop\Mail\Courier:: */
        public function testEmptyMember() {}

        /** @depends \Shop\Mail\Courier::<public> */
        public function testSelector() {}

        /** @depends documented */
        public function testDocumented() {}

        /** @dataProvider abstractProvider */
        public function testAbstractProvider($v) {}

        /** @dataProvider emptyProvider */
        public function testEmptyProvider($v) {}

        /** @dataProvider bareReturn */
        public function testBareReturn($v) {}

        /** @dataProvider interpolated */
        public function testInterpolated($v) {}

        /** @dataProvider holes */
        public function testHoles($v) {}

        /**
         * Seeds first.
         */
        public function testSeeds() {}

        /** documented, not a test */
        public function documented() {}
        abstract public static function abstractProvider();
        public static function emptyProvider() {}
        public static function bareReturn() { return; }
        public static function interpolated() { $k = 'a'; return ["{$k}1" => [1]]; }
        public static function holes() { return [, [1]]; }
    }

    class AnonHolder {
        public function make() {
            return new class {
                /** @depends helper */
                public function testInAnonymous() {}

                /** @covers ::dispatch */
                public function testCoversInAnonymous() {}
            };
        }
    }

    class Broken {
        /** @test */
        public function () {}
    }
}
