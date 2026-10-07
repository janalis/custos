<?php

final class ArchiveTest
{
    public function testGone()
    {
        $this-><weak_warning descr="assertFileNotExists() was deprecated by PHPUnit 9.1; call assertFileDoesNotExist() instead.">AssertFileNotExists</weak_warning>($this->file);
        self::assertnotequals(1, 2, '', <weak_warning descr="PHPUnit 8.0 deprecated the delta argument; call assertNotEqualsWithDelta() instead.">0.5</weak_warning>);
    }
}
