<?php
class Box {}
abstract class BoxTest
{
    abstract protected function box(): Box;
    public function testOld()
    {
        $this->assertInstanceOf(Box::class, $this->box());
    }
}
