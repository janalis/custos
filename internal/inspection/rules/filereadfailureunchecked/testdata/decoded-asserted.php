<?php
namespace App\Tests;

use PHPUnit\Framework\TestCase;

// An assertion that fails on the value a corrupt text decodes to checks the
// read: the failure value decodes like any corrupt text.
final class DecodedTest extends TestCase
{
    public function testDecoded(string $path): void
    {
        $this->assertSame(['a' => 1], unserialize(file_get_contents($path)));
        $this->assertEquals(unserialize(file_get_contents($path)), [1]);
        self::assertIsArray(json_decode(file_get_contents($path), true));
        $this->assertInstanceOf(\stdClass::class, json_decode(file_get_contents($path)));
        $this->assertNotFalse(unserialize(file_get_contents($path)));
        $this->assertNotNull(json_decode(file_get_contents($path)));
        $state = unserialize(file_get_contents($path));
        $this->assertCount(2, $state);
        $this->assertSame([], <warning descr="Handle file reading failure before using its result.">unserialize(file_get_contents($path))</warning>);
        $this->assertEquals(unserialize(file_get_contents($path)), actual: [1]);
        $this->assertEquals($expected, <warning descr="Handle file reading failure before using its result.">unserialize(file_get_contents($path))</warning>);
        $this->assertSame(<warning descr="Handle file reading failure before using its result.">unserialize(file_get_contents($path))</warning>);
        $this->assertSame([1], 'x', <warning descr="Handle file reading failure before using its result.">unserialize(file_get_contents($path))</warning>);
        $this->assertSame(expected: [1], actual: <warning descr="Handle file reading failure before using its result.">unserialize(file_get_contents($path))</warning>);
        $this->assertNotNull(<warning descr="Handle file reading failure before using its result.">unserialize(file_get_contents($path))</warning>);
        $this->assertNotFalse(<warning descr="Handle file reading failure before using its result.">json_decode(file_get_contents($path))</warning>);
        $this->assertFalse(<warning descr="Handle file reading failure before using its result.">unserialize(file_get_contents($path))</warning>);
        $this->assertInstanceOf(<warning descr="Handle file reading failure before using its result.">unserialize(file_get_contents($path))</warning>, $x);
        $this->keep(<warning descr="Handle file reading failure before using its result.">unserialize(file_get_contents($path))</warning>);
        is_callable(<warning descr="Handle file reading failure before using its result.">unserialize(file_get_contents($path))</warning>);
    }

    private function keep($value): void
    {
    }
}
