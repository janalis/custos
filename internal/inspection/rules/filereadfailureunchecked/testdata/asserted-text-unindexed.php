<?php
namespace App\Tests;

use PHPUnit\Framework\TestCase;

// PHPUnit is installed outside the analyzed tree.
final class OutputTest extends TestCase
{
    public function testOutput($stream): void
    {
        $this->assertEquals(escapeshellcmd('done'), escapeshellcmd(stream_get_contents($stream)));
        static::assertSame('done', trim(stream_get_contents($stream)));
        $this->assertSame('', <warning descr="Handle file reading failure before using its result.">trim(stream_get_contents($stream))</warning>);
        $this->check(<warning descr="Handle file reading failure before using its result.">trim(stream_get_contents($stream))</warning>);
    }

    private function check(string $text): void
    {
    }
}
