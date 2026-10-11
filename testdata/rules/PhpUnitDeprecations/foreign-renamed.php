<?php

namespace PHPUnit\Framework {
    abstract class Assert
    {
        public static function assertFileNotExists(string $filename, string $message = ''): void {}
        public static function assertDirectoryNotExists(string $directory, string $message = ''): void {}
    }
    abstract class TestCase extends Assert {}
}

namespace Depot\Storage {
    final class Bucket
    {
        public function assertFileNotExists(string $key): void {}
        public static function assertDirectoryNotExists(string $prefix): void {}

        public function drop(string $key): void
        {
            $this->assertFileNotExists($key);
            self::assertDirectoryNotExists($key);
            static::assertDirectoryNotExists($key);
        }
    }

    final class BucketTest extends \PHPUnit\Framework\TestCase
    {
        public function testDrop(Bucket $bucket, $path): void
        {
            $bucket->assertFileNotExists('report.csv');
            $bucket?->assertFileNotExists('report.csv');
            Bucket::assertDirectoryNotExists('exports');
            $unknown->assertFileNotExists($path);
            $this->make()->assertDirectoryNotExists($path);
            $this-><weak_warning descr="assertFileNotExists() was deprecated by PHPUnit 9.1; call assertFileDoesNotExist() instead.">assertFileNotExists</weak_warning>($path);
            self::<weak_warning descr="assertDirectoryNotExists() was deprecated by PHPUnit 9.1; call assertDirectoryDoesNotExist() instead.">assertDirectoryNotExists</weak_warning>($path);
            \PHPUnit\Framework\Assert::<weak_warning descr="assertFileNotExists() was deprecated by PHPUnit 9.1; call assertFileDoesNotExist() instead.">assertFileNotExists</weak_warning>($path);
            $case = $this;
            $case-><weak_warning descr="assertFileNotExists() was deprecated by PHPUnit 9.1; call assertFileDoesNotExist() instead.">assertFileNotExists</weak_warning>($path);
            $dyn::<weak_warning descr="assertDirectoryNotExists() was deprecated by PHPUnit 9.1; call assertDirectoryDoesNotExist() instead.">assertDirectoryNotExists</weak_warning>($path);
        }
    }
}
