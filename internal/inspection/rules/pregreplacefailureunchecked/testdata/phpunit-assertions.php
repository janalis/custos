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
            $rendered = preg_replace('/a/', 'b', $path);
            $this->assertStringContainsString('Total', $rendered);
            self::assertStringContainsString('Total', preg_replace('/a/', 'b', $path));
            static::assertStringEqualsFile(__FILE__, preg_replace('/a/', 'b', $path));
            Assert::assertStringContainsString('Total', preg_replace('/a/', 'b', $path));
            assertStringContainsString('Total', preg_replace('/a/', 'b', $path));
            $this->expectOutputString(preg_replace('/a/', 'b', $path));
        }

        public function testProjectMethodsUseTheValue($path)
        {
            <warning descr="Handle regex replacement failure before using its result.">$this->createRequest(preg_replace('/a/', 'b', $path))</warning>;
            <warning descr="Handle regex replacement failure before using its result.">self::assertInvoice(preg_replace('/a/', 'b', $path))</warning>;
            <warning descr="Handle regex replacement failure before using its result.">$this->setName(preg_replace('/a/', 'b', $path))</warning>;
        }
    }

    function checkedFromOutside(TestCase $case, $path)
    {
        $case->assertStringContainsString('Total', preg_replace('/a/', 'b', $path));
    }
}
