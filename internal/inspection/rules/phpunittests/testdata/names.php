<?php
namespace Suite;

function count($items) { return 0; }

class NamesTest
{
    public function testNames($bag, $flag, $path)
    {
        <weak_warning descr="Use 'assertCount()' instead.">$this->ASSERTSAME(2, \Count($bag))</weak_warning>;
        $this->assertSame(2, count($bag));
        <weak_warning descr="Use 'assertIsBool()' instead.">self::assertTRUE(IS_BOOL($flag))</weak_warning>;
        // A closed resource is no resource to is_resource(), but is one
        // to assertIsResource() / assertIsNotResource().
        $this->assertFalse(is_resource($path));
        self::assertTrue(\is_resource($path));
        <weak_warning descr="Use 'assertFileExists()' instead.">$this->AssertNotFalse(File_Exists($path))</weak_warning>;
        $stub = $this->createMock(\ArrayAccess::class);
        $stub->Expects(<weak_warning descr="Use '->once()' instead.">$this->Exactly(1)</weak_warning>)->method('offsetGet');
        <weak_warning descr="Use '->willReturn()' instead.">$stub->method('offsetGet')->Will($this->ReturnValue(3))</weak_warning>;
    }
}
