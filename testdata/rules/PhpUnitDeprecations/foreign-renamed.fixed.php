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
            $this->assertFileDoesNotExist($path);
            self::assertDirectoryDoesNotExist($path);
            \PHPUnit\Framework\Assert::assertFileDoesNotExist($path);
            $case = $this;
            $case->assertFileDoesNotExist($path);
            $dyn::assertDirectoryDoesNotExist($path);
        }
    }
}
