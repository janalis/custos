<?php
namespace PHPUnit\Framework {
    abstract class Assert
    {
        public static function assertSame(mixed $expected, mixed $actual, string $message = ''): void {}
        public static function assertTrue(mixed $condition, string $message = ''): void {}
    }
    abstract class TestCase extends Assert {}

    function assertSame(mixed $expected, mixed $actual, string $message = ''): void {}
}

namespace App\Tests {
    use PHPUnit\Framework\TestCase;

    // A read only described in an assertion's failure message.
    final class ProcTest extends TestCase
    {
        public function testRun($pipes, int $status): void
        {
            $output = stream_get_contents($pipes[1]);
            $error = stream_get_contents($pipes[2]);
            $this->assertSame(0, $status, $output.$error);
            $this->assertSame(0, $status, 'Output: '.($output.$error));
            $this->assertTrue($status === 0, message: $output.$error);
            self::assertSame(0, $status, $output.$error);
            \PHPUnit\Framework\assertSame(0, $status, $output.$error);
            $this->assertSame(<warning descr="Handle file reading failure before using its result.">$output.$error</warning>, 'x');
            $this->assertSame(0, $status, $output.$error, <warning descr="Handle file reading failure before using its result.">$error.$output</warning>);
            $this->assertSame(0, $status, ...[<warning descr="Handle file reading failure before using its result.">$output.$error</warning>]);
            $this->assertSame(0, $status, (<warning descr="Handle file reading failure before using its result.">$output.$error</warning>) + 1);
            $this->assertTrue($status === 0, other: <warning descr="Handle file reading failure before using its result.">$output.$error</warning>);
            $this->assertMissing(0, $status, <warning descr="Handle file reading failure before using its result.">$output.$error</warning>);
            echo <warning descr="Handle file reading failure before using its result.">$output.$error</warning>;
        }
    }
}
