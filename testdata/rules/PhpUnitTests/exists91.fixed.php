<?php
class StorageTest extends \PHPUnit\Framework\TestCase
{
    public function testPaths($file, $dir)
    {
        $this->assertFileDoesNotExist($file);
        $this->assertDirectoryDoesNotExist($dir, 'gone');
        $this->assertFileExists($file);
    }
}
