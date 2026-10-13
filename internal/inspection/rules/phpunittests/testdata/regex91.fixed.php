<?php
class SlugTest extends \PHPUnit\Framework\TestCase
{
    public function testSlug($slug)
    {
        $this->assertMatchesRegularExpression('/^[a-z-]+$/', $slug);
        $this->assertDoesNotMatchRegularExpression('/\s/', $slug);
        $this->assertMatchesRegularExpression('/^a/', $slug, 'prefix');
        $this->assertDoesNotMatchRegularExpression('/\d/', $slug);
    }
}
