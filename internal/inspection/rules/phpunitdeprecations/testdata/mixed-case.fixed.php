<?php

final class ArchiveTest
{
    public function testGone()
    {
        $this->assertFileDoesNotExist($this->file);
        self::assertnotequals(1, 2, '', 0.5);
    }
}
