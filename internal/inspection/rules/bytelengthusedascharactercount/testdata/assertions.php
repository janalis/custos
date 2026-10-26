<?php
namespace PHPUnit\Framework {
    abstract class Assert
    {
        public static function assertSame(mixed $expected, mixed $actual, string $message = ""): void {}
        public static function assertCount(int $expected, mixed $haystack, string $message = ""): void {}
    }
    abstract class TestCase extends Assert {}
}

namespace App\Tests {
    use PHPUnit\Framework\TestCase;

    final class MaskTest extends TestCase
    {
        public function testBytes(string $masked): void
        {
            $source = "author: 用户";
            $this->assertSame(\strlen($source), \strlen($masked));
            self::assertSame(strpos($masked, ':'), (strlen('é:')));
            $this->assertSame(<warning descr="Count Unicode characters with the intended counting model.">strlen('é')</warning>, mb_strlen($masked));
            $this->assertCount(<warning descr="Count Unicode characters with the intended counting model.">strlen('é')</warning>, [strlen($masked)]);
            $this->check(<warning descr="Count Unicode characters with the intended counting model.">strlen('é')</warning>, strlen($masked));
            $this->assertSame(...[<warning descr="Count Unicode characters with the intended counting model.">strlen('é')</warning>, strlen($masked)]);
            $this->assertSame('a', 'b', <warning descr="Count Unicode characters with the intended counting model.">strlen('é')</warning>);
            $this->assertSame(<warning descr="Count Unicode characters with the intended counting model.">strlen('é')</warning>, ...$masked);
            $this->assertSame(<warning descr="Count Unicode characters with the intended counting model.">strlen('é')</warning>);
        }

        private function check(int $a, int $b): void {}
    }
}
