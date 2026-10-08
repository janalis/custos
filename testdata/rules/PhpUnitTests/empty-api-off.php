<?php
class OffTest
{
    public function testOff(array $ids)
    {
        <weak_warning descr="Use 'assertNotTrue()' instead.">self::assertTrue(!empty($ids))</weak_warning>;
        $this->assertTrue(empty($ids));
    }
}
