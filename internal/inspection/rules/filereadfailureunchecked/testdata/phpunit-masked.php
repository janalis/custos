<?php
// Assertions that pass when the read failed, as if it gave empty text.
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
    use function PHPUnit\Framework\assertEquals;

    final class ReportTest extends TestCase
    {
        public function testMaskingAssertions($path, $text)
        {
            <warning descr="Handle file reading failure before using its result.">$this->assertStringContainsString(file_get_contents($path), $text)</warning>;
            <warning descr="Handle file reading failure before using its result.">self::assertStringContainsStringIgnoringCase(file_get_contents($path), $text)</warning>;
            <warning descr="Handle file reading failure before using its result.">$this->assertStringStartsWith(file_get_contents($path), $text)</warning>;
            <warning descr="Handle file reading failure before using its result.">$this->assertStringEndsWith(suffix: file_get_contents($path), string: $text)</warning>;
            <warning descr="Handle file reading failure before using its result.">$this->assertEquals("", file_get_contents($path))</warning>;
            <warning descr="Handle file reading failure before using its result.">$this->assertEquals(file_get_contents($path), null)</warning>;
            <warning descr="Handle file reading failure before using its result.">$this->assertEqualsCanonicalizing(null, file_get_contents($path))</warning>;
            <warning descr="Handle file reading failure before using its result.">$this->assertEqualsIgnoringCase("0", file_get_contents($path))</warning>;
            <warning descr="Handle file reading failure before using its result.">$this->assertEmpty(file_get_contents($path))</warning>;
            <warning descr="Handle file reading failure before using its result.">assertEquals(false, file_get_contents($path))</warning>;
        }

        public function testCheckingAssertions($path, $text)
        {
            $this->assertStringContainsString("Total", file_get_contents($path));
            $this->assertStringNotContainsString(file_get_contents($path), $text);
            $this->assertEquals("Total", file_get_contents($path));
            $this->assertSame("", file_get_contents($path));
        }
    }
}
