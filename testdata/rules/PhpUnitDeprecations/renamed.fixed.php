<?php

final class ExportTest
{
    public function testCleanup()
    {
        $this->assertFileDoesNotExist($this->path);
        static::assertDirectoryDoesNotExist(dirname($this->path), 'still there');
        $this->assertFileDoesNotExist($this->path);
    }
}
