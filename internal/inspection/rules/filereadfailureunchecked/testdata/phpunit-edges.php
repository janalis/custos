<?php
// Bindings an assertion's arguments may hide, and operands that are no literal.
namespace PHPUnit\Framework {
    abstract class Assert
    {
        public static function assertStringContainsString(string $needle, string $haystack, string $message = ""): void {}
        public static function assertStringContainsStringIgnoringCase(string $needle, string $haystack, string $message = ""): void {}
        public static function assertStringStartsWith(string $prefix, string $string, string $message = ""): void {}
        public static function assertStringEndsWith(string $suffix, string $string, string $message = ""): void {}
        public static function assertEquals(mixed $expected, mixed $actual, string $message = ""): void {}
        public static function assertEqualsCanonicalizing(mixed $expected, mixed $actual, string $message = ""): void {}
        public static function assertEqualsIgnoringCase(mixed $expected, mixed $actual, string $message = ""): void {}
        public static function assertEmpty(mixed $actual, string $message = ""): void {}
        public static function assertSame(mixed $expected, mixed $actual, string $message = ""): void {}
        public static function assertStringNotContainsString(string $needle, string $haystack, string $message = ""): void {}
    }
    abstract class TestCase extends Assert {}
    function assertEquals(mixed $expected, mixed $actual, string $message = ""): void {}
}

namespace Shop\Tests {
    use PHPUnit\Framework\TestCase;

    final class EdgeTest extends TestCase
    {
        public function testEdges($path, $text, array $options, $expected)
        {
            $this->assertEquals('', file_get_contents($path), ...$options);
            $this->assertEquals($expected, file_get_contents($path));
            <warning descr="Handle file reading failure before using its result.">$this->assertEquals(0.0, file_get_contents($path))</warning>;
            $this->assertEquals(0x0, file_get_contents($path));
            <warning descr="Handle file reading failure before using its result.">$this->assertEquals(actual: file_get_contents($path), expected: FALSE)</warning>;
            $this->assertEquals('');
            $this->assertEquals(actual: '');
            $this->assertEmpty(...$options);
            <warning descr="Handle file reading failure before using its result.">$this->assertStringContainsString(haystack: $text, needle: file_get_contents($path))</warning>;
        }
    }
}
