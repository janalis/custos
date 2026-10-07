<?php
class StorageTest extends \PHPUnit\Framework\TestCase
{
    public function testPaths($file, $dir)
    {
        <weak_warning descr="Use 'assertFileDoesNotExist()' instead.">$this->assertFalse(file_exists($file))</weak_warning>;
        <weak_warning descr="Use 'assertDirectoryDoesNotExist()' instead.">$this->assertNotTrue(is_dir($dir), 'gone')</weak_warning>;
        <weak_warning descr="Use 'assertFileExists()' instead.">$this->assertTrue(file_exists($file))</weak_warning>;
    }
}
