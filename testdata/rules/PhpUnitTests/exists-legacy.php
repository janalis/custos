<?php
class StorageTest extends \PHPUnit\Framework\TestCase
{
    public function testPaths($file, $dir)
    {
        <weak_warning descr="Use 'assertFileNotExists()' instead.">$this->assertFalse(file_exists($file))</weak_warning>;
        <weak_warning descr="Use 'assertDirectoryNotExists()' instead.">$this->assertNotTrue(is_dir($dir))</weak_warning>;
        <weak_warning descr="Use 'assertDirectoryExists()' instead.">$this->assertNotFalse(is_dir($dir))</weak_warning>;
    }
}
