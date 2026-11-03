<?php
namespace PHPUnit\Framework {
    abstract class TestCase
    {
        public static function assertSame($expected, $actual): void {}
    }
}

namespace App\Tests {
    use PHPUnit\Framework\TestCase;

    trait DateAssertions
    {
        public static function assertSameDay(\DateTime $a, \DateTime $b): void
        {
            TestCase::assertSame($a->format('Y-m-d'), $b->format('Y-m-d'));
        }
    }

    abstract class BaseTest extends TestCase
    {
        protected static function assertRedirect(string $url): void {}
    }

    // A test case's own static assert… helpers follow PHPUnit's convention.
    final class DumperTest extends BaseTest
    {
        use DateAssertions;

        public function testDump(): void
        {
            $this->assertSameDay(new \DateTime(), new \DateTime());
            $this->assertRedirect('/');
            $this->assertGenerated('a.php');
            <warning descr="Static method dumpFixture() called through $this; use self::dumpFixture().">$this</warning>->dumpFixture();
        }

        private static function assertGenerated(string $file): void {}

        private static function dumpFixture(): void {}
    }

    // Outside a test case, the same names are ordinary static methods.
    final class Checker
    {
        public function run(): void
        {
            <warning descr="Static method assertValid() called through $this; use self::assertValid().">$this</warning>->assertValid();
        }

        private static function assertValid(): void {}
    }

    final class LooseTest extends \Vendor\MissingCase
    {
        public function testRun(): void
        {
            <warning descr="Static method assertLoose() called through $this; use self::assertLoose().">$this</warning>->assertLoose();
        }

        private static function assertLoose(): void {}
    }
}
