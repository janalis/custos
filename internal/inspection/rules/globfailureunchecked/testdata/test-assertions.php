<?php
namespace PHPUnit\Framework {
    abstract class Assert
    {
        /** @phpstan-assert !false $actual */
        final public static function assertNotFalse(mixed $actual, string $message = ''): void {}
    }

    abstract class TestCase extends Assert {}
}

namespace App\Tests {
    use PHPUnit\Framework\TestCase;

    final class MigrationTest extends TestCase
    {
        // A failing PHPUnit assertion throws: glob() succeeded below it.
        public function testFiles(): void
        {
            $files = glob('/tmp/*.sql');
            static::assertNotFalse($files);
            foreach ($files as $file) {
                echo $file;
            }
        }
    }
}
