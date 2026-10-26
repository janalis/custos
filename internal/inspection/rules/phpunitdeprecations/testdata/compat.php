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
            return $this-><weak_warning descr="assertDirectoryNotExists() was deprecated by PHPUnit 9.1; call assertDirectoryDoesNotExist() instead.">assertDirectoryNotExists</weak_warning>($file);
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
            $this->assertEquals($expected, $actual, '', .5, <weak_warning descr="PHPUnit 8.0 deprecated the maxDepth argument; drop it.">3</weak_warning>);
        }
        if (method_exists($this, 'assertEqualsCanonicalizing') || method_exists($this, $name) || method_exists($this) || method_exists($this, 'x' . 'y')) {
            $this->assertEquals($expected, $actual, '', <weak_warning descr="PHPUnit 8.0 deprecated the delta argument; call assertEqualsWithDelta() instead.">.5</weak_warning>);
        }
        $this->assertEquals($expected, $actual, '', <weak_warning descr="PHPUnit 8.0 deprecated the delta argument; call assertEqualsWithDelta() instead.">.5</weak_warning>);
    }
}
