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
        public function testRollback() {}

        /** @covers \Shop\Billing\Journal */
        public function testJournal() {}

        /** @covers ::strlen */
        public function testMeasures() {}

        /** @covers Book::<protected> */
        public function testInternals() {}

        /** Unlike @covers Nowhere, this is prose. */
        public function audit_notes() {}

        /** @test */
        public function seedsLedger() {}

        /** */
        public function testBalances() {}

        /** @depends seedsLedger */
        public function testAfterSeed() {}

        /** @depends testBalances */
        public function testAfterBalance() {}

        /** @depends helper */
        public function testNeedsHelper() {}

        /** @depends \LedgerTest::vanished */
        public function testNeedsVanished() {}

        /** @dataProvider amounts */
        public function testAmounts($v) {}

        /** @dataProvider plainAmounts */
        public function testPlainAmounts($v) {}

        /** @dataProvider noAmounts */
        public function testNoAmounts($v) {}

        /** @dataProvider phantom */
        public function testPhantom($v) {}

        public function helper() {}
        public static function amounts() { return ['ten' => [10], 'zero' => [0]]; }
        public static function plainAmounts() { return [[10], [0]]; }
        public static function noAmounts() { return array(); }
    }
}
