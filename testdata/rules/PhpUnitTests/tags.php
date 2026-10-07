<?php
namespace Shop\Billing {
    class Ledger extends \SplObjectStorage {
        public function settle() {}
    }
}

namespace {
    use Shop\Billing\Ledger as Book;

    class LedgerTest {
        /** @covers Book::settle */
        public function testSettles() {}

        /** @covers Book::attach() */
        public function testAttachesInherited() {}

        /**
         * Rolls back.
         * @covers Book::rollback
         */
        public function <error descr="The @covers target 'Book::rollback' cannot be resolved.">testRollback</error>() {}

        /** @covers \Shop\Billing\Journal */
        public function <error descr="The @covers target '\Shop\Billing\Journal' cannot be resolved.">testJournal</error>() {}

        /** @covers ::strlen */
        public function testMeasures() {}

        /** @covers Book::<protected> */
        public function testInternals() {}

        /** Unlike @covers Nowhere, this is prose. */
        public function audit_notes() {}

        /** @test */
        public function seedsLedger() {}

        /** <weak_warning descr="Remove '@test': the method name already marks it as a test.">@test</weak_warning> */
        public function testBalances() {}

        /** @depends seedsLedger */
        public function testAfterSeed() {}

        /** @depends testBalances */
        public function testAfterBalance() {}

        /** @depends helper */
        public function <error descr="The @depends target is missing or is not a test.">testNeedsHelper</error>() {}

        /** @depends \LedgerTest::vanished */
        public function <error descr="The @depends target is missing or is not a test.">testNeedsVanished</error>() {}

        /** @dataProvider amounts */
        public function testAmounts($v) {}

        /** @dataProvider plainAmounts */
        public function <weak_warning descr="Give the provider's datasets string keys.">testPlainAmounts</weak_warning>($v) {}

        /** @dataProvider noAmounts */
        public function testNoAmounts($v) {}

        /** @dataProvider phantom */
        public function <error descr="The @dataProvider target cannot be resolved to a method.">testPhantom</error>($v) {}

        public function helper() {}
        public static function amounts() { return ['ten' => [10], 'zero' => [0]]; }
        public static function plainAmounts() { return [[10], [0]]; }
        public static function noAmounts() { return array(); }
    }
}
