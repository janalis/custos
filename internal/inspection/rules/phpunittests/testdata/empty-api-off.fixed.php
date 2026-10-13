<?php
class OffTest
{
    public function testOff(array $ids)
    {
        self::assertNotTrue(empty($ids));
        $this->assertTrue(empty($ids));
    }
}
