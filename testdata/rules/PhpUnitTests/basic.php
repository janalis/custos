<?php
class CartTest
{
    public function testThings($rows, $bag, $cart, $path, $body, $slug)
    {
        <weak_warning descr="Use 'assertNotTrue()' instead.">$this->assertTrue(!($cart->open && $rows))</weak_warning>;
        <weak_warning descr="Use 'assertSame()' instead.">self::assertFalse($cart->size !== 4, 'size')</weak_warning>;
        <weak_warning descr="Use 'assertNotEquals()' instead.">$this->assertTrue($slug <> 'home')</weak_warning>;
        <weak_warning descr="Use 'assertNotEmpty()' instead.">$this->assertFalse(empty($rows))</weak_warning>;
        <weak_warning descr="Use 'assertNull()' instead.">$this->assertSame($cart->coupon, NULL, 'coupon')</weak_warning>;
        <weak_warning descr="Use 'assertIsNotString()' instead.">$this->assertNotTrue(is_string($slug))</weak_warning>;
        <weak_warning descr="Use 'assertInstanceOf()' instead.">$this->assertNotFalse($cart instanceof \Shop\Cart, 'type')</weak_warning>;
        <weak_warning descr="Use 'assertNotInstanceOf()' instead.">$this->assertNotEquals('Shop\\Order', get_class($cart))</weak_warning>;
        <weak_warning descr="Use 'assertDirectoryNotExists()' instead.">$this->assertNotTrue(is_dir($path))</weak_warning>;
        <weak_warning descr="Use 'assertNotCount()' instead.">$this->assertNotEquals(3, count($bag), 'bag')</weak_warning>;
        <weak_warning descr="Use 'assertContains()' instead.">$this->assertNotFalse(in_array($slug, $rows, true))</weak_warning>;
        <weak_warning descr="Use 'assertNotRegExp()' instead.">$this->assertFalse(preg_match('/^[a-z]+$/', $slug) > 0)</weak_warning>;
        <weak_warning descr="Use 'assertRegExp()' instead.">$this->assertNotEquals(0, preg_match('/\d/', $slug), 'digits')</weak_warning>;
        <weak_warning descr="Use 'assertFileEquals()' instead.">$this->assertEquals(file_get_contents($path), file_get_contents($body))</weak_warning>;
        <weak_warning descr="Use 'assertStringEqualsFile()' instead.">$this->assertSame(file_get_contents($path), $body)</weak_warning>;

        $this->assertSame(count($bag), 3);
        $this->assertEquals($body, file_get_contents($path));
        $this->assertTrue(in_array($slug));
        <weak_warning descr="Use 'assertNotTrue()' instead.">$this->AssertTrue(!$rows)</weak_warning>;
        $this->assert(!$rows);
    }

    public function assertStringEqualsFile($file, $text)
    {
        $this->assertSame(file_get_contents($file), $text);
    }

    public function testMocks()
    {
        $stub = $this->createMock(\ArrayAccess::class);
        $stub->expects(<weak_warning descr="Use '->once()' instead.">$this->exactly(1)</weak_warning>)->method('offsetGet');
        <weak_warning descr="Use '->willReturnCallback()' instead.">$stub->method('offsetGet')->will($this->returnCallback('strtoupper'))</weak_warning>;
        $stub->expects($this->exactly(2))->method('offsetSet');
        $stub->method('offsetExists')->will($this->returnSelf());
    }
}
