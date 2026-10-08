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
        // A closed resource is no resource to is_resource(), but is one
        // to assertIsResource() / assertIsNotResource().
        $this->assertFalse(is_resource($path));
        self::assertTrue(\is_resource($path));
        $this->assertFileExists($path);
        $stub = $this->createMock(\ArrayAccess::class);
        $stub->Expects($this->once())->method('offsetGet');
        $stub->method('offsetGet')->willReturn(3);
    }
}
