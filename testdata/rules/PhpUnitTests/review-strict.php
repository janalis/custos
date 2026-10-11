<?php
declare(strict_types=1);

class StrictTest
{
    public function testStrict($a, $cls, $body, int $n, string $path, array $rows, $pet)
    {
        <weak_warning descr="Use 'assertCount()' instead.">$this->assertEquals($n, count($rows))</weak_warning>;
        <weak_warning descr="Use 'assertCount()' instead.">$this->assertEquals($a, count($rows))</weak_warning>;
        <weak_warning descr="Use 'assertInstanceOf()' instead.">$this->assertTrue($pet instanceof $cls)</weak_warning>;
        <weak_warning descr="Use 'assertStringEqualsFile()' instead.">$this->assertSame(file_get_contents($path), $body)</weak_warning>;
    }
}
