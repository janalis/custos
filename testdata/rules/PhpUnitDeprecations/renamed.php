<?php

final class ExportTest
{
    public function testCleanup()
    {
        $this-><weak_warning descr="assertFileNotExists() was deprecated by PHPUnit 9.1; call assertFileDoesNotExist() instead.">assertFileNotExists</weak_warning>($this->path);
        static::<weak_warning descr="assertDirectoryNotExists() was deprecated by PHPUnit 9.1; call assertDirectoryDoesNotExist() instead.">assertDirectoryNotExists</weak_warning>(dirname($this->path), 'still there');
        $this->assertFileDoesNotExist($this->path);
    }
}
