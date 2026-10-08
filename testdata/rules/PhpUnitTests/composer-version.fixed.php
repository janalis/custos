<?php
// PHPUnit is not indexed: the version comes from composer.json (^11.5).
use PHPUnit\Framework\TestCase;

final class IdsTest extends TestCase
{
    public function testIds(): void
    {
        $ids = ['1', '2'];
        $this->assertTrue(in_array(1, $ids));
        $this->assertMatchesRegularExpression('/a/', 'abc');
    }
}
