<?php
class StorageTest extends \PHPUnit\Framework\TestCase
{
    public function testPaths($file, $dir)
    {
        $this->assertFileNotExists($file);
        $this->assertDirectoryNotExists($dir);
        $this->assertDirectoryExists($dir);
    }
}
