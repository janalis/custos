<?php

namespace PHPUnit\Framework {
    abstract class Assert
    {
        public static function assertStringContainsString(string $needle, string $haystack, string $message = ''): void
        {
        }

        public static function assertStringEqualsFile(string $expectedFile, string $actualString, string $message = ''): void
        {
        }
    }

    abstract class TestCase extends Assert
    {
        public function expectOutputString(string $expected): void
        {
        }

        public function setName(string $name): void
        {
        }
    }

    function assertStringContainsString(string $needle, string $haystack, string $message = ''): void
    {
    }
}

namespace Shop\Tests {
    use PHPUnit\Framework\Assert;
    use PHPUnit\Framework\TestCase;

    use function PHPUnit\Framework\assertStringContainsString;

    abstract class InvoiceTestCase extends TestCase
    {
        protected function createRequest(string $body): array
        {
            return [$body];
        }

        protected static function assertInvoice(string $rendered): void
        {
        }
    }

    final class InvoiceTest extends InvoiceTestCase
    {
        public function testAssertionsCheckTheValue($path)
        {
            $rendered = file_get_contents($path);
            $this->assertStringContainsString('Total', $rendered);
            self::assertStringContainsString('Total', file_get_contents($path));
            static::assertStringEqualsFile(__FILE__, file_get_contents($path));
            Assert::assertStringContainsString('Total', file_get_contents($path));
            assertStringContainsString('Total', file_get_contents($path));
            $this->expectOutputString(file_get_contents($path));
        }

        public function testProjectMethodsUseTheValue($path)
        {
            <warning descr="Handle file reading failure before using its result.">$this->createRequest(file_get_contents($path))</warning>;
            <warning descr="Handle file reading failure before using its result.">self::assertInvoice(file_get_contents($path))</warning>;
            <warning descr="Handle file reading failure before using its result.">$this->setName(file_get_contents($path))</warning>;
        }
    }

    function checkedFromOutside(TestCase $case, $path)
    {
        $case->assertStringContainsString('Total', file_get_contents($path));
    }
}
