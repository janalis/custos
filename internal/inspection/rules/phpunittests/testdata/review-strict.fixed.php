<?php
declare(strict_types=1);

class StrictTest
{
    public function testStrict($a, $cls, $body, int $n, string $path, array $rows, $pet)
    {
        $this->assertCount($n, $rows);
        $this->assertEquals($a, count($rows));
        $this->assertTrue($pet instanceof $cls);
        $this->assertSame(file_get_contents($path), $body);
    }
}
