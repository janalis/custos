<?php
class LayoutTest
{
    public function testLayout($cart, $path)
    {
        <weak_warning descr="Use 'assertFileEquals()' instead.">$this->assertStringEqualsFile(
            $path,
            file_get_contents($cart->path()),
        )</weak_warning>;
        <weak_warning descr="Use 'assertSame()' instead.">$this->assertTrue(
            $cart->size === 4
        )</weak_warning>;
        <weak_warning descr="Use 'assertNull()' instead.">$this->assertSame(
            $cart->coupon, null)</weak_warning>;
    }
}
