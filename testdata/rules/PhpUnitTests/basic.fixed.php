<?php
class CartTest
{
    public function testThings(array $rows, $bag, $cart, $path, $body, $slug)
    {
        $this->assertNotTrue($cart->open && $rows);
        self::assertSame($cart->size, 4, 'size');
        $this->assertNotEquals($slug, 'home');
        $this->assertNotEmpty($rows);
        $this->assertNull($cart->coupon, 'coupon');
        $this->assertIsNotString($slug);
        $this->assertInstanceOf(\Shop\Cart::class, $cart, 'type');
        $this->assertNotInstanceOf(\Shop\Order::class, $cart);
        $this->assertDirectoryNotExists($path);
        $this->assertNotCount(3, $bag, 'bag');
        $this->assertContains($slug, $rows);
        $this->assertNotRegExp('/^[a-z]+$/', $slug);
        $this->assertRegExp('/\d/', $slug, 'digits');
        $this->assertFileEquals($path, $body);
        $this->assertStringEqualsFile($path, $body);

        $this->assertSame(count($bag), 3);
        $this->assertEquals($body, file_get_contents($path));
        $this->assertTrue(in_array($slug));
        $this->assertNotTrue($rows);
        $this->assert(!$rows);
    }

    public function assertStringEqualsFile($file, $text)
    {
        $this->assertSame(file_get_contents($file), $text);
    }

    public function testMocks()
    {
        $stub = $this->createMock(\ArrayAccess::class);
        $stub->expects($this->once())->method('offsetGet');
        $stub->method('offsetGet')->willReturnCallback('strtoupper');
        $stub->expects($this->exactly(2))->method('offsetSet');
        $stub->method('offsetExists')->will($this->returnSelf());
    }
}
