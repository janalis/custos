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
        public function <error descr="The @covers target '::nowhere' cannot be resolved.">testAttributed</error>() {}

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
        public function <error descr="The @covers target '\Shop\Mail\deliver' cannot be resolved.">testFunctionAsClass</error>() {}

        /** @covers \Shop\Mail\Missing */
        public function <error descr="The @covers target '\Shop\Mail\Missing' cannot be resolved.">testMissingClass</error>() {}

        /** @dataProvider ::provider */
        public function <error descr="The @dataProvider target cannot be resolved to a method.">testEmptyClassPart</error>($v) {}

        /** @depends \Shop\Mail\Courier:: */
        public function <error descr="The @depends target is missing or is not a test.">testEmptyMember</error>() {}

        /** @depends \Shop\Mail\Courier::<public> */
        public function <error descr="The @depends target is missing or is not a test.">testSelector</error>() {}

        /** @depends documented */
        public function <error descr="The @depends target is missing or is not a test.">testDocumented</error>() {}

        /** @dataProvider abstractProvider */
        public function testAbstractProvider($v) {}

        /** @dataProvider emptyProvider */
        public function testEmptyProvider($v) {}

        /** @dataProvider bareReturn */
        public function testBareReturn($v) {}

        /** @dataProvider interpolated */
        public function testInterpolated($v) {}

        /** @dataProvider holes */
        public function <weak_warning descr="Give the provider's datasets string keys.">testHoles</weak_warning>($v) {}

        /**
         * Seeds first.
         * <weak_warning descr="Remove '@test': the method name already marks it as a test.">@test</weak_warning>
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
                public function <error descr="The @depends target is missing or is not a test.">testInAnonymous</error>() {}

                /** @covers ::dispatch */
                public function <error descr="The @covers target '::dispatch' cannot be resolved.">testCoversInAnonymous</error>() {}
            };
        }
    }

    class Broken {
        /** @test */
        public function () {}
    }
}
