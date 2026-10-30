<?php
class LayoutTest
{
    public function testLayout($cart, $path)
    {
        $this->assertFileEquals(
            $path,
            $cart->path(),
        );
        $this->assertSame(
            4,
            $cart->size
        );
        $this->assertNull($cart->coupon);
    }
}
