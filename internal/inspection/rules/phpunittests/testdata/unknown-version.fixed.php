<?php
// Neither PHPUnit nor its version is known: assertContains() compares
// strictly from 9.0, unlike in_array(), so it is not suggested. The regex
// and negative resource names changed in 9.1 and the old ones are gone
// from 10, so those are reported, naming both, without a fix.
use PHPUnit\Framework\TestCase;

final class TagsTest extends TestCase
{
    public function testTags($tag, array $tags): void
    {
        $this->assertTrue(in_array($tag, $tags));
        $this->assertNotTrue(in_array($tag, $tags), 'absent');
    }

    public function testNames(string $s, string $f): void
    {
        $this->assertSame(1, preg_match('/a/', $s));
        $this->assertFalse(preg_match('/a/', $s) > 0);
        $this->assertFalse(file_exists($f));
        $this->assertDirectoryExists($f);
    }
}
