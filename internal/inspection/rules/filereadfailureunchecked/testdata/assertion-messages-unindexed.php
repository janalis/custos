<?php
namespace App\Tests;

use PHPUnit\Framework\TestCase;

// PHPUnit installed outside the analyzed tree: common assertions' message
// positions are known.
final class ProcTest extends TestCase
{
    public function testRun($pipes, int $status): void
    {
        $output = stream_get_contents($pipes[1]);
        $error = stream_get_contents($pipes[2]);
        $this->assertSame(0, $status, $output.$error);
        $this->assertTrue($status === 0, $output.$error);
        $this->assertSame(0, <warning descr="Handle file reading failure before using its result.">$output.$error</warning>);
        $this->assertCustom(0, $status, <warning descr="Handle file reading failure before using its result.">$output.$error</warning>);
    }
}
