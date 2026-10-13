<?php
// PHPUnit is not indexed: the version comes from composer.json (^11.5).
use PHPUnit\Framework\TestCase;

final class IdsTest extends TestCase
{
    public function testIds($a, $b): void
    {
        $ids = ['1', '2'];
        $this->assertTrue(in_array(1, $ids));
        <weak_warning descr="Use 'assertSame()' instead.">$this->assertTrue($a === $b)</weak_warning>;
        <weak_warning descr="Use 'assertMatchesRegularExpression()' instead.">$this->assertSame(1, preg_match('/a/', 'abc'))</weak_warning>;
    }
}
