<?php
use PHPUnit\Framework\TestCase;

final class PartialTest extends TestCase
{
    public function testPartials(): void
    {
        // A trailing `...` passes the remaining arguments later: no delta.
        $check = $this->assertEquals(?, ?, ?, ...);
        // A named placeholder binds by name.
        $fold = $this->assertEquals(1, ?, '', <weak_warning descr="PHPUnit 8.0 deprecated the ignoreCase argument; call assertEqualsIgnoringCase() instead.">ignoreCase: ?</weak_warning>);
    }
}
