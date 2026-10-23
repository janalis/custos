<?php
namespace PHPUnit\Framework {
    abstract class Assert
    {
        /** @phpstan-assert !false $actual */
        final public static function assertNotFalse(mixed $actual, string $message = ''): void {}

        /** @phpstan-assert resource $actual */
        final public static function assertIsResource(mixed $actual, string $message = ''): void {}
    }

    abstract class TestCase extends Assert {}
}

namespace App\Tests {
    use PHPUnit\Framework\TestCase;

    final class Check
    {
        /** @phpstan-assert !false $value */
        public static function notFalse(mixed $value): void {}
    }

    final class StreamTest extends TestCase
    {
        // A failing PHPUnit assertion throws: the stream is open below it.
        public function testResource(): void
        {
            $stream = fopen('php://memory', 'rb+');
            static::assertIsResource($stream);
            fwrite($stream, 'x');
        }

        public function testNotFalse(): void
        {
            $stream = fopen('php://memory', 'rb+');
            $this->assertNotFalse($stream);
            fwrite($stream, 'x');
        }

        // A project's own assertion is only documented.
        public function testOwnCheck(): void
        {
            $stream = fopen('php://memory', 'rb+');
            Check::notFalse($stream);
            <warning descr="Check that the stream opened successfully.">fwrite($stream, 'x')</warning>;
        }
    }
}
