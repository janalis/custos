<?php
namespace Suite;

function count($items) { return 0; }

class NamesTest
{
    public function testNames($bag, $flag, $path)
    {
        $this->assertCount(2, $bag);
        $this->assertSame(2, count($bag));
        self::assertIsBool($flag);
        $this->assertFileExists($path);
        $stub = $this->createMock(\ArrayAccess::class);
        $stub->Expects($this->once())->method('offsetGet');
        $stub->method('offsetGet')->willReturn(3);
    }
}
