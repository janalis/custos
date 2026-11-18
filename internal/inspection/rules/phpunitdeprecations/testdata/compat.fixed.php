<?php

trait PolyfillAssert
{
    public static function assertDirectoryDoesNotExist($directory, $message = '')
    {
        static::assertDirectoryNotExists($directory, $message);
    }

    public static function assertFileDoesNotExist($file, $message = '')
    {
        static::assertFileNotExists($file, $message);
        $check = function () use ($file) {
            return $this->assertDirectoryDoesNotExist($file);
        };
    }
}

final class DurationTest
{
    public function testDuration($expected, $actual)
    {
        if (method_exists($this, 'assertEqualsWithDelta')) {
            $this->assertEqualsWithDelta($expected, $actual, .5);
        } else {
            $this->assertEquals($expected, $actual, '', .5);
        }
        \method_exists($this, 'ASSERTFILEDOESNOTEXIST') ? null : $this->assertFileNotExists('x');
        if ($ready) {
        } elseif (!method_exists($this, 'assertEqualsWithDelta')) {
            $this->assertEquals($expected, $actual, '', .5, 3);
        }
        if (method_exists($this, 'assertEqualsCanonicalizing') || method_exists($this, $name) || method_exists($this) || method_exists($this, 'x' . 'y')) {
            $this->assertEquals($expected, $actual, '', .5);
        }
        $this->assertEquals($expected, $actual, '', .5);
    }
}
