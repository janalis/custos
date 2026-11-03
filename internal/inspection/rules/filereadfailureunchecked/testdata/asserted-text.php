<?php
namespace PHPUnit\Framework {
    abstract class Assert
    {
        public static function assertEquals(mixed $expected, mixed $actual, string $message = ''): void {}
        public static function assertSame(mixed $expected, mixed $actual, string $message = ''): void {}
        public static function assertEmpty(mixed $actual, string $message = ''): void {}
    }
    abstract class TestCase extends Assert {}

    function assertSame(mixed $expected, mixed $actual, string $message = ''): void {}
}

namespace App\Tests {
    use PHPUnit\Framework\TestCase;

    // A read turned into text that keeps the empty text, then asserted.
    final class OutputTest extends TestCase
    {
        public function testOutput($stream, string $expected): void
        {
            $this->assertEquals(escapeshellcmd('done'), escapeshellcmd(stream_get_contents($stream)));
            $this->assertSame('done', trim(strtolower((stream_get_contents($stream)))));
            \PHPUnit\Framework\assertSame('done', nl2br(stream_get_contents($stream)));
            self::assertSame($expected, ucfirst(string: stream_get_contents($stream)));
            $this->assertSame('', <warning descr="Handle file reading failure before using its result.">trim(stream_get_contents($stream))</warning>);
            $this->assertEquals(null, <warning descr="Handle file reading failure before using its result.">trim(stream_get_contents($stream))</warning>);
            $this->assertEmpty(<warning descr="Handle file reading failure before using its result.">trim(stream_get_contents($stream))</warning>);
            $this->assertSame('done', <warning descr="Handle file reading failure before using its result.">str_repeat(stream_get_contents($stream), 2)</warning>);
            $this->assertSame('done', <warning descr="Handle file reading failure before using its result.">trim('x', stream_get_contents($stream))</warning>);
            $this->assertSame('done', <warning descr="Handle file reading failure before using its result.">trim(stream_get_contents($stream))</warning> . 'x');
            $this->assertSame('done', mine(<warning descr="Handle file reading failure before using its result.">trim(stream_get_contents($stream))</warning>));
            $this->assertSame('done', strrev(<warning descr="Handle file reading failure before using its result.">trim(stream_get_contents($stream))</warning>));
        }
    }

    function mine(string $text): string
    {
        return $text;
    }
}
