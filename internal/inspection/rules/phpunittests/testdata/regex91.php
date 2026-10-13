<?php
class SlugTest extends \PHPUnit\Framework\TestCase
{
    public function testSlug($slug)
    {
        <weak_warning descr="Use 'assertMatchesRegularExpression()' instead.">$this->assertSame(1, preg_match('/^[a-z-]+$/', $slug))</weak_warning>;
        <weak_warning descr="Use 'assertDoesNotMatchRegularExpression()' instead.">$this->assertNotSame(1, preg_match('/\s/', $slug))</weak_warning>;
        <weak_warning descr="Use 'assertMatchesRegularExpression()' instead.">$this->assertTrue(preg_match('/^a/', $slug) > 0, 'prefix')</weak_warning>;
        <weak_warning descr="Use 'assertDoesNotMatchRegularExpression()' instead.">$this->assertFalse(preg_match('/\d/', $slug) > 0)</weak_warning>;
    }
}
