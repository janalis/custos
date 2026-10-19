<?php
// A null passes loose comparisons and emptiness checks; a string needle rejects it.
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
            $this->assertStringContainsString(preg_replace('/a/', 'b', $path), $text);
            self::assertStringContainsStringIgnoringCase(preg_replace('/a/', 'b', $path), $text);
            $this->assertStringStartsWith(preg_replace('/a/', 'b', $path), $text);
            $this->assertStringEndsWith(suffix: preg_replace('/a/', 'b', $path), string: $text);
            <warning descr="Handle regex replacement failure before using its result.">$this->assertEquals("", preg_replace('/a/', 'b', $path))</warning>;
            <warning descr="Handle regex replacement failure before using its result.">$this->assertEquals(preg_replace('/a/', 'b', $path), null)</warning>;
            <warning descr="Handle regex replacement failure before using its result.">$this->assertEqualsCanonicalizing(null, preg_replace('/a/', 'b', $path))</warning>;
            $this->assertEqualsIgnoringCase("0", preg_replace('/a/', 'b', $path));
            <warning descr="Handle regex replacement failure before using its result.">$this->assertEmpty(preg_replace('/a/', 'b', $path))</warning>;
            <warning descr="Handle regex replacement failure before using its result.">assertEquals(false, preg_replace('/a/', 'b', $path))</warning>;
        }

        public function testCheckingAssertions($path, $text)
        {
            $this->assertStringContainsString("Total", preg_replace('/a/', 'b', $path));
            $this->assertStringNotContainsString(preg_replace('/a/', 'b', $path), $text);
            $this->assertEquals("Total", preg_replace('/a/', 'b', $path));
            $this->assertSame("", preg_replace('/a/', 'b', $path));
        }
    }
}
