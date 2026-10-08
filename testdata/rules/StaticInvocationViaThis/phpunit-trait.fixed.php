<?php
namespace PHPUnit\Framework {
    abstract class Assert {
        final public static function assertIsResource(mixed $actual, string $message = ''): void {}
        public function assertSoon(): void {}
    }
    abstract class TestCase extends Assert {}
}

namespace App\Tests {
    // A trait restating PHPUnit's assertions abstractly, so static analysers
    // narrow through them: the using test case gets PHPUnit's own.
    trait ResourceAsserts {
        abstract public static function assertIsResource(mixed $actual, string $message = ''): void;
        abstract public static function assertSoon(): void;
        abstract public static function assertLoaded(mixed $actual): void;

        public function check($handle): void {
            $this->assertIsResource($handle);
            static::assertSoon();
            static::assertLoaded($handle);
        }
    }
}
