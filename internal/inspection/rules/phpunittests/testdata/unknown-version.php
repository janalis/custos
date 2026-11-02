<?php
// Neither PHPUnit nor its version is known: assertContains() compares
// strictly from 9.0, unlike in_array(), so it is not suggested.
use PHPUnit\Framework\TestCase;

final class TagsTest extends TestCase
{
    public function testTags($tag, array $tags): void
    {
        $this->assertTrue(in_array($tag, $tags));
        $this->assertNotTrue(in_array($tag, $tags), 'absent');
    }
}
